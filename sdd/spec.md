# Spec: API Gateway Lite

## Portfolio Contract

- Project: `#18 api-gateway-lite`
- Macro project: `delivery-observability-infra`
- Claim: authenticate, enforce a shared quota, propagate telemetry context, and quantify gateway overhead.
- Default runtime: local Docker Compose with no paid credentials.

## Functional Requirements

1. Reject a missing or invalid `X-API-Key` with `401` and never forward that credential.
2. Apply a token bucket per authenticated principal through one atomic Redis operation shared by all gateway instances.
3. Return quota headers; return `429` when empty and fail closed with `503` when Redis is unavailable.
4. Preserve a valid `X-Correlation-ID` or generate one, return it, forward it, and attach it to the server span.
5. Extract and inject W3C trace context and export spans through a configurable OTLP/HTTP endpoint.
6. Reverse proxy path, query, headers, method, and body to the configured HTTP(S) upstream.

## Evidence Requirements

- Unit tests for auth, policy middleware, proxying, correlation, and benchmark calculation.
- Real Redis tests proving shared quota and atomic burst consumption across two limiter instances.
- OTLP exporter test against a real HTTP receiver stub.
- Compose smoke contract covering `401`, `200`, `429`, correlation, and trace propagation.
- Benchmark V2 with at least three repetitions and exact provenance.

## Out of Scope

- User identity, OAuth/OIDC, API-key lifecycle, TLS termination, WAF, dynamic routes, service discovery, retries, circuit breakers, caching, or managed-cloud parity.
- Claims beyond one Docker host and the documented echo workload.

## Definition of Done

- [x] Runtime code implements every functional requirement.
- [x] Docker Compose defines gateway, upstream, Redis, and an OTLP collector.
- [x] Benchmark runner rejects dirty or unidentified source state.
- [x] A V2 artifact exists and README opens with its measured result.
- [x] Full local validation passes from the final worktree.
