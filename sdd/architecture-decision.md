# Architecture Decision

## Status

Accepted

## Context

Project: api-gateway-lite
Claim: gateway com auth, rate limit e observabilidade
Benchmark: overhead_ms

Problem forces:

- Domain complexity: low
- Integration pressure: low
- UI state complexity: none
- Data/ML reproducibility: low
- Auditability/event history: low
- Throughput/async pressure: medium
- Independent deployability need: low

## Decision

Chosen architecture: modular-monolith

Reason:

The gateway is a single binary with composable middleware layers (auth, rate limit, tracing) wrapping a reverse proxy. This avoids microservice complexity while keeping each concern independently testable and swappable. Each middleware is a self-contained package with its own tests and a clean `func(http.Handler) http.Handler` interface.

Dependency rule:

Middleware depends on the handler interface only; config is parsed at startup and injected; no domain code depends on transport detail.

## Rejected Alternatives

| Alternative | Why rejected |
|---|---|
| microservices | adds deployment and coordination complexity without benefit for a single-process gateway |
| hexagonal architecture | ports/adapters abstraction adds indirection that inflates overhead_ms benchmark |

## Folder Layout

```
cmd/
  api-gateway-lite/    entry point, parses flags, starts server
  bench-target/        simple echo server for benchmarking
  benchmark/           benchmark runner (starts target+gateway, measures overhead)
internal/
  config.go            env-var-based configuration
  proxy.go             ReverseProxy handler
  auth.go              API key validation middleware
  ratelimit/           token bucket + HTTP middleware
  telemetry/           OTel setup + tracing middleware
  benchmark/           benchmark suite (measure functions)
```

## Testing Strategy

- Unit tests: each middleware and component tested with httptest
- Integration tests: proxy_test.go verifies forwarding and error handling
- Benchmark: in-process servers + measure functions produce overhead_ms

## Consequences

Positive:

- Each concern (auth, rate limit, tracing) is independently testable
- stdlib ReverseProxy is lightweight and battle-tested
- Zero external dependencies for default run path (no Redis, no DB)

Tradeoffs:

- Single-instance only; no distributed rate limiting without Redis
- In-memory token bucket state is lost on restart

Migration path:

- Replace in-memory bucket with Redis-backed counter via interface
- Replace noop tracer with OTLP exporter by changing env var
