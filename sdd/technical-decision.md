# Technical Decision

## Status

Accepted

## Decision Type

stack, library, runtime

## Context

Project: api-gateway-lite
Problem: Minimal gateway with auth, rate limiting, and observability — measured by latency overhead
Portfolio program: delivery-observability-infra
Public signal: Go reverse-proxy middleware proficiency
Benchmark: overhead_ms

## Selected Option

Selected: Go stdlib net/http/httputil.ReverseProxy + in-memory token bucket + OTel noop exporter

Reason:

Go's stdlib ReverseProxy is the lightest possible forwarder — no framework overhead. The in-memory token bucket avoids a Redis dependency for the default local path. OTel with noop exporter gives full tracing API without requiring a collector, keeping the default demo credential-free.

## Decision Brain Fields

- Stack profile: go-backend
- API style: rest-http
- Messaging: none
- Cloud mode: none
- Database/runtime: go 1.22, alpine 3.20
- Library policy: stdlib for proxy, OTel SDK for tracing, in-memory for rate limiting

## Engineering Principles

Coupling boundary:

Middleware depends on `http.Handler` interface; no domain code depends on transport detail.

SOLID application:

- SRP: each internal package owns exactly one concern (auth, rate-limit, proxy, telemetry)
- OCP: middleware stack composable by wrapping handlers; new middleware added without modifying existing code
- LSP: any `http.Handler` can be wrapped; no interface subversion
- ISP: middleware interface is `func(http.Handler) http.Handler` — one method
- DIP: high-level HTTP server depends on `http.Handler` abstraction, not concrete middleware

Simplicity:

- KISS: in-memory token bucket over Redis avoids unnecessary network dependency
- YAGNI: circuit breaker, retry logic, request logging not added
- DRY: duplicated business knowledge removed without premature abstraction

Testability evidence:

- Auth test: httptest without real servers
- Proxy test: httptest upstream without starting real process
- Limiter test: pure Go, no network
- Benchmark test: measure function tested with httptest

## Rejected Options

| Option | Why rejected |
|---|---|
| gin-gonic/gin framework | stdlib is lighter for benchmark purity; gin adds measurable overhead |
| Redis-backed rate limiter | unnecessary network hop for local demo; YAGNI for single-instance |
| OTLP gRPC exporter | stdout exporter is simpler for local demo; gRPC adds dependency weight |

## API Contract

Contract artifact: none (internal proxy, no public API surface)

GraphQL controls, when applicable: N/A

## Cloud Local-First

Local provider: none

Real provider target: none

Config switch: `CLOUD_PROVIDER=none`

Unsupported local behaviors: none

## Benchmark Impact

Expected impact: overhead_ms should be < 2ms for the simple echo path

Validation command: `go run ./cmd/benchmark`

## Operational Cost

- Docker services added: none
- Local demo complexity: low
- Failure case required: no

## Follow-up

- If overhead_ms > 5ms, investigate httputil.ReverseProxy copy overhead
