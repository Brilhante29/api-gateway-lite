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
| Cross-platform benchmark wrappers should derive commit, image, and dependency digests and reject dirty trees. | backlog | benchmark harness | Generalize after this project proves both wrappers. | recorded |
| Redis server-time Lua plus a two-client atomicity contract is reusable for Go rate-limit projects. | backlog | Go backend skill | Extract the contract shape, not project package code. | recorded |
| Compose smoke should verify context at the upstream boundary, not only gateway response status. | backlog | CI templates | Add an optional upstream-observation gate. | recorded |
| This collector configuration is intentionally a debug-only local adapter. | reject | runtime templates | Keep project-specific until another repo needs the same trace-only topology. | resolved |

The local run confirmed all three backlog candidates without changing another repository: the wrappers preserved exact provenance, the Redis contract caught cross-container networking assumptions, and the upstream-observation smoke proved both correlation and W3C trace propagation.

## Scope Control

No kit repository is modified from this worktree. The reusable deltas are documented for the principal portfolio task because changing another repository would violate the single-repository execution constraint.

## Final Gate

- [x] Reusable improvements were patched or recorded.
- [x] Project-specific implementation was not moved into the kit.
- [x] Validation reflects the stale-documentation and weak-provenance mistakes discovered during the project.
