# Reuse Improvement Review

Project: `18 - api-gateway-lite`

## Review Points

- [x] after scaffold
- [x] after architecture decision
- [x] after first working slice
- [x] after benchmark result
- [x] before final local release
- [ ] after CI failure, if applicable

## Findings

| Finding | Classification | Kit area | Resolution | Status |
|---|---|---|---|---|
| Cross-platform benchmark wrappers should derive commit, image, and dependency digests and reject dirty trees. | patch_now | benchmark harness | Both wrappers now preserve canonical evidence locally and accept a separate CI result directory. | resolved |
| Redis server-time Lua plus a two-client atomicity contract is reusable for Go rate-limit projects. | backlog | Go backend skill | Extract the contract shape, not project package code. | recorded |
| Compose smoke should verify context at the upstream boundary, not only gateway response status. | backlog | CI templates | Add an optional upstream-observation gate. | recorded |
| This collector configuration is intentionally a debug-only local adapter. | reject | runtime templates | Keep project-specific until another repo needs the same trace-only topology. | resolved |

The local run confirmed the reusable candidates: the wrappers preserved exact provenance, the Redis contract caught cross-container networking assumptions, and the upstream-observation smoke proved both correlation and W3C trace propagation. The central kit records the stable-publication-versus-CI-smoke rule.

## Scope Control

Project-specific gateway code remains here. Only the general benchmark evidence rule belongs in `portfolio-reuse-kit`.

## Final Gate

- [x] Reusable improvements were patched or recorded.
- [x] Project-specific implementation was not moved into the kit.
- [x] Validation reflects the stale-documentation and weak-provenance mistakes discovered during the project.
