package performance_engine

import (
	"strings"
	"sync"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// LoadTestRunner executes the actual load simulation.
type LoadTestRunner struct {
	httpTool *shared.HTTPTool
	params   PerformanceParams
}

// NewLoadTestRunner creates a new load test runner.
func NewLoadTestRunner(httpTool *shared.HTTPTool, params PerformanceParams) *LoadTestRunner {
	if params.Concurrency <= 0 {
		params.Concurrency = 10
	}
	if params.Duration <= 0 {
		params.Duration = 30
	}
	return &LoadTestRunner{
		httpTool: httpTool,
		params:   params,
	}
}

// Run executes the performance test according to the mode.
func (r *LoadTestRunner) Run(endpoints map[string]shared.EndpointAnalysis) ExecutionMetrics {
	var metricsCollector MetricsCollector
	var wg sync.WaitGroup

	stop := make(chan struct{})
	duration := time.Duration(r.params.Duration) * time.Second

	// Launch virtual users (goroutines)
	for i := 0; i < r.params.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					// Select an endpoint (simplistically pick the first one for now or rotate)
					for epKey := range endpoints {
						metricsCollector.Record(r.executeRequest(epKey))

						// Respect RPS if specified
						if r.params.RPS > 0 {
							time.Sleep(time.Second / time.Duration(r.params.RPS))
						}
					}
				}
			}
		}(i)
	}

	// Wait for duration
	time.Sleep(duration)
	close(stop)
	wg.Wait()

	return metricsCollector.Finalize()
}

func (r *LoadTestRunner) executeRequest(epKey string) RequestStat {
	start := time.Now()

	method, path := "GET", "/"
	if parts := strings.SplitN(epKey, " ", 2); len(parts) == 2 {
		method, path = parts[0], parts[1]
	}

	resp, err := r.httpTool.Run(shared.HTTPRequest{
		Method: method,
		URL:    r.params.BaseURL + path,
	})
	if err != nil {
		return RequestStat{Latency: time.Since(start), Success: false}
	}
	return RequestStat{
		StatusCode: resp.StatusCode,
		Latency:    resp.Duration,
		Success:    resp.StatusCode < 500,
	}
}

type RequestStat struct {
	StatusCode int
	Latency    time.Duration
	Success    bool
}
