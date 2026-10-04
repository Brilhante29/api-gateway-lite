# API Gateway Lite: Authentication, Shared Quotas, and Trace Propagation in Go

**`21.533 ms` median p95 overhead and `1,617.13 req/s` through the gateway**, with zero rejects and zero failures across three Docker repetitions. A small Go gateway enforces API-key authentication and one atomic Redis quota across replicas, and propagates correlation IDs and W3C trace context to any HTTP upstream.

[![CI](https://github.com/Brilhante29/api-gateway-lite/actions/workflows/ci.yml/badge.svg)](https://github.com/Brilhante29/api-gateway-lite/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white) ![Redis](https://img.shields.io/badge/Redis-7.4-DC382D?logo=redis&logoColor=white) ![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-425CC7?logo=opentelemetry&logoColor=white)

## Why this exists

Every service eventually needs the same edge concerns: who is calling, how much they may call, and how to follow one request across services when something breaks. Full API-management platforms solve this with a lot of moving parts; hand-rolled middleware tends to leak credentials upstream, keep quotas per replica, or drop the trace at the hop that matters. This gateway does the essential part, measured:

- constant-time API-key checks, with credentials stripped before proxying;
- one token bucket per key in Redis, refilled and consumed atomically by a Lua script using Redis `TIME`, so replicas share the quota without trusting local clocks;
- `429` with standard and compatibility rate-limit headers when the quota is exhausted, and `503` (fail closed) when Redis is down;
- validated correlation IDs and W3C trace context propagated upstream and exported over OTLP/HTTP.

## Quickstart

```bash
docker compose up --build --wait
curl -i -H "X-API-Key: local-demo-key" -H "X-Correlation-ID: demo-1" http://localhost:8080/echo
docker compose logs otel-collector   # inspect exported traces
docker compose down --volumes
```

The response is `200 ok` with quota headers and `X-Correlation-ID`, and it shows the correlation and `traceparent` values the upstream observed. No paid credential is required.

## How it works

```text
client
  -> correlation + inbound OTel span
  -> constant-time API-key check
  -> atomic Redis token bucket
  -> instrumented reverse proxy (httputil.ReverseProxy)
  -> upstream with correlation + W3C trace context
  -> OTLP/HTTP collector
```

`TELEMETRY=none` disables export without changing request policy.

## Results

The harness compares the same `/echo` upstream directly and through the full gateway path. It alternates measurement order, warms both paths, runs three repetitions, and records latency percentiles, overhead, throughput, rejects, failures, workload digests, image digest, dependency-lock digest, exact commit, and producer.

| Metric | Median of 3 runs | Unit |
|---|---:|---|
| `overhead_p50_ms` | 7.363 | ms |
| `overhead_p95_ms` | 21.533 | ms |
| `overhead_p99_ms` | 30.896 | ms |
| `gateway_p95_ms` | 23.559 | ms |
| `direct_throughput_rps` | 17,749.46 | requests/second |
| `gateway_throughput_rps` | 1,617.13 | requests/second |
| `gateway_rejects` | 0 | requests |
| `direct_failures` / `gateway_failures` | 0 / 0 | requests |

How to read it: every gateway request pays for authentication, a Redis round trip for the quota decision, span creation and export, and a second HTTP hop, against a tiny echo payload on one Docker host. The overhead is the price of those guarantees in this setup; it is not an internet or multi-region capacity figure. Compare only artifacts with the same `comparability_key`.

Evidence source: clean commit `10371288ad7b400fb6b73dcaf9c1f0a680df0345`, measured with Go `1.23.12` on Linux/amd64 (Docker Desktop); the runtime has since moved to Go 1.26 for security fixes. The committed JSON keeps unrounded samples and all provenance digests. Regenerate with `sh ./tools/run-benchmark.sh` (refuses a dirty worktree; writes `benchmarks/results/latest.json`).

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `API_KEY` | `local-demo-key` | Demo credential; inject a secret in real deployments |
| `UPSTREAM_URL` | `http://localhost:8081` | HTTP(S) upstream |
| `REDIS_ADDR` | `localhost:6379` | Shared quota store |
| `RATE_LIMIT` | `100` | Tokens added per second |
| `RATE_BURST` | `200` | Bucket capacity |
| `TELEMETRY` | `otlp` | `otlp` or `none` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTel SDK default | OTLP/HTTP collector endpoint |

## Design decisions

| Decision | Why | Rejected |
|---|---|---|
| `net/http` + `httputil.ReverseProxy` | Standard library, small surface, easy to audit | A full gateway framework for four concerns |
| Redis Lua token bucket with server time | One quota across replicas, no clock-skew tokens | Per-replica in-memory limits |
| Fail closed on Redis outage | Protected traffic must not bypass the quota | Fail open |
| OpenTelemetry with W3C propagation | Vendor-neutral traces through any backend | Proprietary tracing headers |

## Limitations

- API keys are a focused mechanism, not user identity, OAuth, key rotation, TLS termination, or an authorization server.
- One upstream URL; no control plane, service discovery, retries, circuit breaker, cache, WAF, or body policy.
- The local collector uses a debug exporter; production backends plug in through the OTLP endpoint.

## Testing

```bash
go test -race ./...
go vet ./...
```

CI repeats formatting, dependency-lock checks, vet, race tests, real Redis contract tests, Compose smoke checks, benchmark V2 generation, and artifact validation.

## Project structure

```text
cmd/api-gateway-lite/   gateway entrypoint
cmd/bench-target/       echo upstream used by the benchmark
cmd/benchmark/          direct-versus-gateway harness
internal/               auth, config, gateway, proxy, ratelimit, telemetry, benchmark
deploy/                 OpenTelemetry Collector configuration
benchmarks/  tools/     results, runners, validators
sdd/  openspec/         decisions, benchmark plan, handoff
```

## How this repository is built

The project follows the spec-driven workflow of [portfolio-reuse-kit](https://github.com/Brilhante29/portfolio-reuse-kit). Requirements and decisions live in [`sdd/`](sdd) and [`openspec/`](openspec), and [`project.yaml`](project.yaml) records the architecture, stack, and rejected alternatives. Development is AI-assisted and human-governed: [`AGENTS.md`](AGENTS.md) and [`CLAUDE.md`](CLAUDE.md) hold the coding-agent instructions, while tests, validators, and CI decide what gets published.

## Related work

- [go-rate-limiter](https://github.com/Brilhante29/go-rate-limiter): the distributed token bucket in isolation, benchmarked across two nodes.
- [observability-stack](https://github.com/Brilhante29/observability-stack): correlating one incident across metrics, traces, and logs.

See [`REFERENCES.md`](REFERENCES.md) for sources and reuse attribution.

## Author

**Guilherme Brilhante**, software engineer working on scalable backends and production AI.
[LinkedIn](https://www.linkedin.com/in/guilhermefreirebrilhanteseveriano/) · [GitHub](https://github.com/Brilhante29) · [Publications](https://dblp.org/pid/353/6812.html)

## License

[MIT](LICENSE).
