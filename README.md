# #18 api-gateway-lite

**Status:** benchmarked

**Claim:** gateway com auth, rate limit e observabilidade.

**Benchmark target:** overhead_ms — latency overhead introduced by the gateway.

**Stack:** go, reverse-proxy, opentelemetry, docker.

## Benchmark

| Metric | Value | Unit |
|---|---:|---:|
| overhead_ms | 0.34 | ms |

Run the benchmark locally:

```bash
go run ./cmd/benchmark
```

Or via Docker:

```bash
docker build -t api-gateway-lite .
docker run --rm api-gateway-lite benchmark
```

## Run

Start the target upstream server:

```bash
go run ./cmd/bench-target
```

Then start the gateway (default :8080 → :8081):

```bash
export API_KEY=my-secret
go run ./cmd/api-gateway-lite
```

Test with curl:

```bash
curl -H "X-API-Key: my-secret" http://localhost:8080/echo
```

## Architecture

```
client → auth middleware → rate limiter → tracing middleware → reverse proxy → upstream
```

Each middleware layer is independently testable. The reverse proxy uses Go's stdlib `httputil.ReverseProxy`.

## References

See REFERENCES.md.
