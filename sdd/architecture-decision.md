# Architecture Decision

## Decision

Use a modular monolith: one Go process composes authentication, quota, telemetry, and reverse-proxy modules around `http.Handler`. Keep external behavior behind the narrow `ratelimit.Limiter` and `HealthChecker` capabilities; construct Redis and OpenTelemetry adapters only at startup.

## Why This Fits

The problem has one synchronous request lifecycle, low domain complexity, high integration pressure, and a low-latency benchmark. Independent deployment of each concern would add hops and failure modes without creating independently valuable capabilities.

```text
cmd/api-gateway-lite (composition root)
  -> internal/auth
  -> internal/ratelimit.Limiter <- Redis Lua adapter
  -> internal/telemetry        <- OTLP/HTTP adapter
  -> internal/proxy            -> arbitrary HTTP upstream
```

Dependency direction points from the composition root to small policies and ports. Redis clients, environment reads, and exporter construction do not leak into authentication, proxy, or benchmark policy.

## Principles

- SRP: auth, quota, telemetry, proxy, configuration, and evidence each have one owner.
- OCP: another limiter or exporter can be injected without rewriting gateway composition.
- LSP: limiter implementations must preserve allow/reject/error semantics; real Redis tests define the current contract.
- ISP: `Limiter.Allow` and `HealthChecker.Ping` expose only the capabilities their callers need.
- DIP: the handler pipeline depends on interfaces; startup selects concrete infrastructure.
- KISS/YAGNI: standard `net/http`, one upstream, one Redis script, and OTLP are enough for the claim.
- DRY: quota arithmetic exists once in the atomic script; propagation and evidence schemas have one implementation each.

## Rejected Alternatives

| Alternative | Reason |
|---|---|
| Microservices | Adds network hops and coordination to one request policy pipeline. |
| Full clean-architecture rings | No independent business entity or use case justifies the extra layers. |
| In-memory limiter | Breaks shared quota under replicas and restart. |
| Gateway framework | Hides the standard HTTP costs this repository measures. |

## Consequences

Redis availability is now part of protected request availability, intentionally enforced as fail-closed. OTLP failure does not change authorization or quota behavior because export is asynchronous. Dynamic routing and resilience policies remain outside this repository.
