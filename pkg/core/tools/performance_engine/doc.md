# Performance Engine Module

## Overview
The Performance Engine is a high-performance load testing tool that supports multiple testing modes and provides high-resolution latency metrics.

## Tools
- `run_performance`: Executes load, stress, spike, or soak tests.

## Testing Modes
- **Load**: Validate performance under expected traffic levels.
- **Stress**: Find the breaking point of the API.
- **Spike**: Test response to sudden traffic bursts.
- **Soak**: Identify memory leaks and stability issues over time.

## Metrics
- Average Latency
- Percentiles: p50, p95, p99
- Success/Failure distribution
- Throughput (RPS)

## Usage
```json
{
  "mode": "stress",
  "base_url": "http://localhost:3000",
  "concurrency": 50,
  "duration_sec": 60
}
```
