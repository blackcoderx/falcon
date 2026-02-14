package smoke_runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// runHealthChecks executes reachability and basic functionality checks.
func (t *SmokeRunnerTool) runHealthChecks(baseURL string, endpoints []string) []HealthCheck {
	var checks []HealthCheck

	for _, ep := range endpoints {
		parts := strings.SplitN(ep, " ", 2)
		method := "GET"
		path := ep
		if len(parts) == 2 {
			method = parts[0]
			path = parts[1]
		}

		url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

		check := HealthCheck{
			Name:     fmt.Sprintf("%s %s", method, path),
			Endpoint: url,
		}

		start := time.Now()
		req := shared.HTTPRequest{
			Method: method,
			URL:    url,
		}

		resp, err := t.httpTool.Run(req)
		check.Latency = time.Since(start).String()

		if err != nil {
			check.Status = "error"
			check.Message = err.Error()
		} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			check.Status = "ok"
			check.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		} else {
			check.Status = "error"
			check.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}

		checks = append(checks, check)
	}

	return checks
}
