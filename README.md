# #18 API Gateway Lite

**Claim:** a small Go gateway can enforce API-key authentication and one atomic Redis quota across replicas while propagating correlation and W3C trace context to any HTTP upstream.

**Measured result:** `21.533 ms` median p95 overhead and `1,617.13 req/s` through the gateway, with zero rejects and zero failures across three Docker repetitions.

**Stack:** Go 1.23, `net/http`, `httputil.ReverseProxy`, Redis 7.4, OpenTelemetry OTLP/HTTP, OpenTelemetry Collector, Docker Compose.

## Run

```bash
docker compose up --build --wait
curl -i -H "X-API-Key: local-demo-key" -H "X-Correlation-ID: demo-1" http://localhost:8080/echo
```

The response is `200 ok`, includes quota headers, returns `X-Correlation-ID`, and exposes the correlation and `traceparent` values observed by the upstream. No paid credential is required. Stop the stack with:

```bash
docker compose down --volumes
```

## What It Proves

```text
client
  -> correlation + inbound OTel span
  -> constant-time API-key check
  -> atomic Redis token bucket
  -> instrumented reverse proxy
  -> upstream with correlation + W3C trace context
  -> OTLP/HTTP collector
```

- Authentication credentials are removed before proxying.
- A Lua script uses Redis `TIME` and updates refill plus consumption atomically, so replicas share one quota without trusting local clocks.
- The gateway fails closed with `503` when Redis is unavailable and emits `429` plus standard and compatibility rate-limit headers when quota is exhausted.
- Correlation IDs are validated, returned to the caller, forwarded upstream, and attached to the server span.
- OpenTelemetry uses standard W3C propagation and exports over OTLP/HTTP; `TELEMETRY=none` disables export without changing request policy.

Inspect local trace exports with `docker compose logs otel-collector`.

## Benchmark V2

The harness compares the same `/echo` upstream directly and through the production gateway path. It alternates measurement order, warms both paths, runs three repetitions, and records p50/p95/p99 latency, overhead, throughput, rejects, failures, workload digests, image digest, dependency-lock digest, exact commit, and producer.

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

Evidence source: clean commit `10371288ad7b400fb6b73dcaf9c1f0a680df0345`, Go `1.23.12`, Linux/amd64 on Docker Desktop. The committed JSON retains unrounded samples and all provenance digests.

```powershell
pwsh ./tools/run-benchmark.ps1
```

```bash
sh ./tools/run-benchmark.sh
```

The runner intentionally refuses a dirty worktree. Output: `benchmarks/results/latest.json`.
CI sends regenerated smoke evidence to `runner.temp`; it never replaces this committed publication baseline.

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

## Limits

- API keys are a focused gateway mechanism, not user identity, OAuth, key rotation, TLS termination, or an authorization server.
- Routes are configured through one upstream URL; there is no control plane, service discovery, retry, circuit breaker, cache, WAF, or request-body policy.
- Redis is mandatory because shared quota is part of the claim. Its outage deliberately rejects protected traffic.
- Benchmark numbers measure a tiny echo payload on one Docker host. Compare only artifacts with the same `comparability_key`; they are not internet or multi-region capacity claims.
- The local collector uses a debug exporter. Production backends remain pluggable through the OTLP endpoint and their own authentication settings.

## Verify

```powershell
./tools/validate-project.ps1
```

CI repeats formatting, dependency-lock, vet, race tests, real Redis contract tests, Compose smoke checks, benchmark V2 generation, and artifact validation. Design decisions live in `sdd/`; sources and reuse attribution live in `REFERENCES.md`.
