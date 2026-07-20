# Reuse Improvement Review

Project: `18 - api-gateway-lite`

## Review Points

- [x] after scaffold
- [x] after architecture decision
- [x] after first working slice
- [x] after benchmark result
- [x] before publication
- [ ] after CI failure, if applicable

## Findings

| Finding | Classification | Kit Area | Action | Status |
|---|---|---|---|---|
| Token bucket rate limiter is generic enough for kit reuse | backlog | harness | Extract to portfolio-reuse-kit/templates/ratelimit | recorded |
| Benchmark suite pattern (in-process servers) could be templated | backlog | templates | Extract in-process benchmark pattern | recorded |

## Patch Now Decisions

- None (project-specific implementation; no kit fix needed)

## Backlog Decisions

- Token bucket rate limiter template for future Go projects
- In-process benchmark harness template

## Rejected Improvements

| Improvement | Reason |
|---|---|
| OTel span attributes in semconv format | Adds dependency coupling; YAGNI for current benchmark scope |
| Config file support (YAML/TOML) | Env vars suffice for the 6-config surface area |

## Final Gate

- [x] Reusable improvements were patched or recorded.
- [x] Project-specific implementation was not moved into the kit.
- [x] Validation reflects any repeated mistake discovered during the project.
