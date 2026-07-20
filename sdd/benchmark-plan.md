# Benchmark Plan: api-gateway-lite

## Hypothesis

gateway com auth, rate limit e observabilidade, measured by overhead_ms — the latency difference between a direct upstream call and a gateway-mediated call.

## Command

```bash
go run ./cmd/benchmark
```

## Environment

- OS: any (Docker or Go 1.22+)
- CPU: any
- RAM: any
- GPU: none
- Docker version: any
- Date: recorded in result JSON

## Inputs

- fixture: in-process echo server
- dataset size: N/A
- repetitions: 100 requests per path
- warmup: none

## Metrics

| Metric | Unit | Source | Why it matters |
|---|---:|---|---|
| overhead_ms | ms | benchmark suite | proves the repo claim — the gateway should add minimal latency |

## Result schema

Output must be JSON and include project, metric, value, unit, timestamp, environment, and command. Written to `benchmarks/results/latest.json`.

## Post angle

#18 api-gateway-lite: overhead_ms as a reproducible portfolio benchmark — a minimal Go gateway that proves auth + rate limiting + tracing can stay under measurable overhead.
