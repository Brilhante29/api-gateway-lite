# References and Reuse

| Reference | License | Applied idea | Copied code |
|---|---|---|---|
| Go `net/http/httputil` | BSD-3-Clause | Standard reverse-proxy implementation and HTTP interoperability | no |
| Redis command and Lua scripting documentation | RSALv2/SSPLv1 for Redis distribution; documentation terms apply | Server-time atomic token-bucket state transition | no |
| `redis/go-redis` | BSD-2-Clause | Redis client and script execution API | no |
| OpenTelemetry Go | Apache-2.0 | Server/client instrumentation, W3C propagation, OTLP export | no |
| OpenTelemetry Collector Contrib | Apache-2.0 | Credential-free local OTLP receiver and debug exporter | no |
| `portfolio-reuse-kit` local snapshot | repository license | SDD, decision graph, Go profile, benchmark V2 contract, and release gates | patterns/templates only |

All gateway code, Lua script, tests, Compose topology, fixture, benchmark runner, and result are project-specific. No external source code was copied into this repository.
