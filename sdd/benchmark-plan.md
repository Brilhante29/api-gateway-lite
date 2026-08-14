# Benchmark Plan: Gateway Overhead V2

## Hypothesis

The gateway adds measurable but bounded latency to the same echo workload while preserving successful throughput with zero rejects and failures under a deliberately non-binding benchmark quota.

## Reproducible Command

```powershell
pwsh ./tools/run-benchmark.ps1
```

```bash
sh ./tools/run-benchmark.sh
```

Both wrappers require a clean commit, build one immutable local image, capture its digest and the `go.sum` digest, then run the same image as upstream, gateway, and benchmark client under Compose.

## Workload

- Fixture: `GET /echo`, response `200 ok`.
- Warmup: 200 requests per path.
- Measured: 2,000 requests per path per repetition.
- Concurrency: 16 workers.
- Repetitions: 3 minimum.
- Order control: direct then gateway on even runs; gateway then direct on odd runs.
- Gateway path: API-key auth, Redis Lua quota, correlation middleware, OTel server/client instrumentation, OTLP export, reverse proxy.

## Metrics

| Metric family | Unit | Direction |
|---|---|---|
| Direct and gateway p50/p95/p99 | ms | lower |
| Overhead p50/p95/p99 | ms | lower |
| Direct and gateway throughput | requests/second | higher |
| Direct/gateway failures and gateway rejects | requests | lower |

The primary public metric is median `overhead_p95_ms` across repetitions. The artifact also records every repetition, summaries, workload/config digests, runtime, architecture, exact source commit, clean-tree gate, image and dependency digests, producer, command, and comparability key.

## Interpretation Limits

The direct and gateway runs are sequential within each repetition, so host scheduling can produce noise or occasional negative deltas. Alternating order reduces systematic bias. Results compare only identical comparability keys on similar Docker hosts and do not model TLS, large bodies, WAN latency, or multi-region Redis.

## Output

`benchmarks/results/latest.json`, conforming to `.portfolio/contracts/benchmark-result-v2.schema.json` and requiring at least three samples per metric.
