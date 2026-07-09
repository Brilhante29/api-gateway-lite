# #18 api-gateway-lite

**Status:** scaffold

**Proves:** gateway com auth, rate limit e observabilidade.

**Benchmark target:** overhead_ms.

**Stack:** go, reverse-proxy, redis, opentelemetry, k6, docker.

## Next milestone

Implement the smallest Docker-runnable version and produce the first JSON benchmark under enchmarks/results/.

## Run

`ash
docker build -t api-gateway-lite .
docker run --rm api-gateway-lite
`

## Benchmark

`ash
docker run --rm api-gateway-lite benchmark
`

| Metric | Value | Unit |
|---|---:|---|
| overhead_ms | pending | pending |

## Architecture

Defined in sdd/spec.md before implementation.

## References

See REFERENCES.md.