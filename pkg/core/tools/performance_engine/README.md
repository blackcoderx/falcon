# Performance Engine (`pkg/core/tools/performance_engine`)

The Performance Engine runs load and stress tests against your API.

## Key Tool: `run_performance`

This tool executes multi-mode performance tests and tracks high-resolution latency metrics.

### Modes

- **load**: Simulates expected traffic volume to verify stability.
- **stress**: Pushes the system beyond its limits to find breaking points.
- **spike**: Sudden bursts of traffic to test resilience.
- **soak**: Long-duration tests to detect memory leaks.

## Metrics

Tracks total requests, success rate, RPS (Requests Per Second), and latency percentiles (p50, p95, p99).

## Example Prompts

Trigger this tool by asking:
- "Run a load test on the API with 50 concurrent users."
- "Stress test the checkout endpoint to find its breaking point."
- "Simulate a traffic spike on the search API."
- "Run a soak test for 1 hour to check for memory leaks."
