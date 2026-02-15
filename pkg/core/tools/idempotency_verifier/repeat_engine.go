package idempotency_verifier

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// RepeatEngine repeats requests and compares responses to detect idempotency issues.
type RepeatEngine struct {
	httpTool    *shared.HTTPTool
	baseURL     string
	repeatCount int
}

// Verify checks a set of endpoints for idempotency violations.
func (e *RepeatEngine) Verify(endpoints map[string]shared.EndpointAnalysis) IdempotencyResult {
	var result IdempotencyResult

	for epKey := range endpoints {
		result.TotalVerified++

		parts := strings.SplitN(epKey, " ", 2)
		method := "POST"
		path := epKey
		if len(parts) == 2 {
			method = parts[0]
			path = parts[1]
		}

		url := strings.TrimSuffix(e.baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

		// 1. Initial request
		req := shared.HTTPRequest{
			Method: method,
			URL:    url,
			// Simplified: In a real implementation we'd use the parameter engine
			// to generate valid data once and repeat the SAME data.
		}

		resp1, err := e.httpTool.Run(req)
		if err != nil {
			continue
		}

		// 2. Repeat requests
		isIdempotent := true
		for i := 1; i < e.repeatCount; i++ {
			respN, err := e.httpTool.Run(req)
			if err != nil {
				isIdempotent = false
				result.Violations = append(result.Violations, Violation{
					Endpoint:    epKey,
					Description: fmt.Sprintf("Repeated request #%d failed: %v", i, err),
				})
				break
			}

			// 3. Compare responses
			// Note: Status codes should be consistent (e.g. 200 or 201 then 200/204)
			// For POST, idempotency often means same status or specific error if duplicate.
			// But for strictly idempotent methods like PUT, it should be identical state.

			if method == "POST" {
				// POST isn't necessarily idempotent unless specified.
				// We check if we get a 201 (Created) again, which MIGHT mean a duplicate was created.
				if resp1.StatusCode == 201 && respN.StatusCode == 201 {
					isIdempotent = false
					result.Violations = append(result.Violations, Violation{
						Endpoint:    epKey,
						Description: "POST returned 201 Created twice; potential duplicate resource creation.",
					})
					break
				}
			} else if resp1.StatusCode != respN.StatusCode {
				isIdempotent = false
				result.Violations = append(result.Violations, Violation{
					Endpoint:    epKey,
					Description: fmt.Sprintf("Response status changed from %d to %d on repetition.", resp1.StatusCode, respN.StatusCode),
				})
				break
			}
		}

		if isIdempotent {
			result.IdempotentCount++
		}
	}

	return result
}
