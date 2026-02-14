package regression_watchdog

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// DiffEngine compares current API responses against baseline snapshots.
type DiffEngine struct {
	httpTool *shared.HTTPTool
	baseURL  string
}

// Check identifies behavioral changes between live API and the baseline.
func (e *DiffEngine) Check(baseline *APIBaseline, filter []string) RegressionResult {
	var result RegressionResult
	result.BaselineDate = baseline.CreatedAt.Format("2006-01-02 15:04:05")

	for epKey, snapshot := range baseline.Snapshots {
		// Filter endpoints if specified
		if len(filter) > 0 {
			match := false
			for _, f := range filter {
				if f == epKey {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		parts := strings.SplitN(epKey, " ", 2)
		method := "GET"
		path := epKey
		if len(parts) == 2 {
			method = parts[0]
			path = parts[1]
		}

		url := strings.TrimSuffix(e.baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

		req := shared.HTTPRequest{
			Method: method,
			URL:    url,
		}

		resp, err := e.httpTool.Run(req)
		if err != nil {
			result.Regressions = append(result.Regressions, Regression{
				Endpoint:    epKey,
				ChangeType:  "error",
				Description: fmt.Sprintf("Failed to reach endpoint: %v", err),
			})
			continue
		}

		// Compare Status Code
		if resp.StatusCode != snapshot.StatusCode {
			result.Regressions = append(result.Regressions, Regression{
				Endpoint:    epKey,
				ChangeType:  "status_code",
				Description: fmt.Sprintf("Status changed from %d to %d", snapshot.StatusCode, resp.StatusCode),
			})
			continue
		}

		// Compare Body structure/content (Simplified for simulation)
		if resp.Body != snapshot.Body {
			// In a real implementation, we'd use a structural JSON diff
			result.Regressions = append(result.Regressions, Regression{
				Endpoint:    epKey,
				ChangeType:  "body_diff",
				Description: "Response body content or structure has changed",
			})
			continue
		}

		result.StableCount++
	}

	return result
}
