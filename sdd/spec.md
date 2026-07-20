# Spec: api-gateway-lite

## Number

#18

## Claim

Este projeto prova que: gateway com auth, rate limit e observabilidade.

## Stack

go, reverse-proxy, opentelemetry, docker

## User-visible output

- Docker command: `docker build -t api-gateway-lite . && docker run --rm api-gateway-lite`
- README opens with: # #18 api-gateway-lite
- Benchmark table: overhead_ms

## Scope

In:

- Implementar o menor produto funcional que prove o claim.
- Rodar por Docker.
- Gerar benchmark JSON reproduzivel.

Out:

- Publicar repo antes do primeiro resultado numerico.
- Depender de segredo pago para o caminho default.

## Architecture

```
client → auth middleware → rate limiter → tracing middleware → reverse proxy → upstream
```

## Benchmark

Primary metric:

- name: overhead_ms
- target: first reproducible baseline
- command: `go run ./cmd/benchmark`
- result file: `benchmarks/results/latest.json`

## Dataset or fixture

- source: self-contained (in-process echo server)
- size: N/A
- license: N/A
- deterministic seed: 42

## Definition of done

- [x] Docker command works from clean clone.
- [x] README starts with project number and benchmark result.
- [x] Benchmark command writes JSON result.
- [x] Tests cover core behavior.
- [x] REFERENCES.md explains reuse.
- [x] No secret or paid credential required for default demo.
