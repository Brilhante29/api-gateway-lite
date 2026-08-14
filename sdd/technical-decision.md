# Technical Decision

## Selected Stack

- Go 1.23 and `net/http`/`httputil.ReverseProxy` for a visible, standard request path.
- `go-redis/v9` with one Lua token-bucket operation using Redis server time.
- OpenTelemetry Go HTTP instrumentation plus OTLP/HTTP exporter.
- Redis 7.4 and OpenTelemetry Collector under Docker Compose.

## Rate-Limit Semantics

The authenticated API key is hashed into a stable principal. The key itself is removed before forwarding. Redis atomically reads server time, refills tokens, consumes one token when available, stores state, and refreshes TTL. The Redis key uses a hash tag around the principal so the script remains single-slot compatible. Adapter errors produce `503`; exhausted quota produces `429` and `Retry-After`.

## Telemetry Semantics

The outer correlation middleware accepts only bounded `[A-Za-z0-9_.-]` values and otherwise creates a cryptographically random ID. The OTel server handler extracts inbound W3C context and starts a span; the outbound transport injects the resulting context. `correlation.id` is a span attribute. OTLP endpoint and backend authentication are SDK environment configuration, keeping vendor SDKs out of request policy.

## API Decision

REST/HTTP is selected because this component transparently forwards arbitrary HTTP upstreams. GraphQL would impose a schema and resolver lifecycle unrelated to the claim; gRPC would narrow interoperability and belongs in project #15. The gateway-owned contract is limited to `GET /healthz`, `X-API-Key`, `X-Correlation-ID`, W3C trace headers, and rate-limit response headers.

## Local-First and Cloud

The default path uses local Redis and an OTel Collector and needs no secret. Kumo is not started because this repository does not model an AWS service. A real observability backend plugs in through OTLP environment variables; no cloud SDK is imported.

## Rejected Libraries and Infrastructure

| Option | Reason |
|---|---|
| Gin/Fiber | Standard `net/http` compatibility and transparent overhead matter more than routing ergonomics. |
| In-memory quota | Cannot prove atomic shared behavior. |
| Kafka/RabbitMQ | The path is synchronous and has no event delivery semantics. |
| Vendor tracing SDK | OTLP already provides the required pluggability. |
| Retry/circuit-breaker library | Changes proxy semantics and benchmark scope without being part of the claim. |
