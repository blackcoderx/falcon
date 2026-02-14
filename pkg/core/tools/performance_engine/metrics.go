package performance_engine

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ExecutionMetrics holds the final results of a performance test run.
type ExecutionMetrics struct {
	Total       int           `json:"total"`
	Success     int           `json:"success"`
	Fail        int           `json:"fail"`
	SuccessRate float64       `json:"success_rate"`
	AvgLatency  time.Duration `json:"avg_latency"`
	P50         time.Duration `json:"p50"`
	P95         time.Duration `json:"p95"`
	P99         time.Duration `json:"p99"`
	Min         time.Duration `json:"min"`
	Max         time.Duration `json:"max"`
	RPS         float64       `json:"rps"`
}

// MetricsCollector accumulates request statistics during a test run.
type MetricsCollector struct {
	mu       sync.Mutex
	stats    []RequestStat
	duration time.Duration
}

// Record adds a single request statistic to the collector.
func (c *MetricsCollector) Record(stat RequestStat) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats = append(c.stats, stat)
}

// Finalize calculates the final metrics from the collected statistics.
func (c *MetricsCollector) Finalize() ExecutionMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.stats) == 0 {
		return ExecutionMetrics{}
	}

	var totalLatency time.Duration
	var successCount int
	latencies := make([]int64, 0, len(c.stats))

	min := c.stats[0].Latency
	max := c.stats[0].Latency

	for _, s := range c.stats {
		totalLatency += s.Latency
		if s.Success {
			successCount++
		}
		latencies = append(latencies, int64(s.Latency))
		if s.Latency < min {
			min = s.Latency
		}
		if s.Latency > max {
			max = s.Latency
		}
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	count := len(latencies)
	metrics := ExecutionMetrics{
		Total:       count,
		Success:     successCount,
		Fail:        count - successCount,
		SuccessRate: float64(successCount) / float64(count) * 100,
		AvgLatency:  totalLatency / time.Duration(count),
		Min:         min,
		Max:         max,
		P50:         time.Duration(latencies[int(float64(count)*0.50)]),
		P95:         time.Duration(latencies[int(float64(count)*0.95)]),
		P99:         time.Duration(latencies[int(float64(count)*0.99)]),
	}

	return metrics
}

// FormatSummary returns a human-readable summary of the performance metrics.
func (m *ExecutionMetrics) FormatSummary(mode string) string {
	res := fmt.Sprintf("🚀 Performance Test Complete (Mode: %s)\n\n", mode)
	res += fmt.Sprintf("Requests:   %d total, %d success, %d failed (%.2f%% success rate)\n", m.Total, m.Success, m.Fail, m.SuccessRate)
	res += fmt.Sprintf("Latency:\n")
	res += fmt.Sprintf("  Avg: %v\n", m.AvgLatency)
	res += fmt.Sprintf("  Min: %v\n", m.Min)
	res += fmt.Sprintf("  Max: %v\n", m.Max)
	res += fmt.Sprintf("Percentiles:\n")
	res += fmt.Sprintf("  p50: %v\n", m.P50)
	res += fmt.Sprintf("  p95: %v\n", m.P95)
	res += fmt.Sprintf("  p99: %v\n", m.P99)
	return res
}
