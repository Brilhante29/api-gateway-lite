# Agent Handoff

## Current Objective

Finish `#18 api-gateway-lite` as a benchmarked local repository. Do not push or use a GitHub token in this worktree.

## Decisions Already Closed

| Role | Decision | Evidence |
|---|---|---|
| Program planner | Keep the repo in `delivery-observability-infra`. | `project.yaml` |
| Architecture selector | Modular monolith with ports at Redis and health boundaries. | `sdd/architecture-decision.md` |
| Principles reviewer | SOLID, LSP, DIP, KISS, YAGNI, and fail-closed semantics are explicit and tested. | architecture ADR and tests |
| Stack/API agents | Go stdlib reverse proxy over REST/HTTP; Redis and official OTel packages. | `sdd/technical-decision.md` |
| Cloud/messaging agents | Local Compose, no Kumo because no AWS behavior, no broker, OTLP backend pluggability. | Compose and technical ADR |
| Benchmark agent | V2 direct-vs-gateway harness with three repetitions and exact provenance. | `internal/benchmark`, scripts |
| Reuse reviewer | Record portable Redis and V2 wrapper patterns for kit extraction; do not couple this repo back to the kit. | reuse review |

## Runtime

- Demo: `docker compose up --build --wait`
- Request: `curl -i -H "X-API-Key: local-demo-key" http://localhost:8080/echo`
- Benchmark: `pwsh ./tools/run-benchmark.ps1` or `sh ./tools/run-benchmark.sh`
- Validation: `./tools/validate-project.ps1`

## Verification State

- Runtime implementation: complete.
- Unit/integration test definitions: complete.
- Docker/Compose and CI definitions: complete.
- Benchmark evidence: pending clean implementation commit and run.
- Final local validation and clean worktree: pending.

## Continuation Rule

Read `AGENTS.md`, then this file and `git status`. Preserve the exact provenance in `benchmarks/results/latest.json`; never hand-edit measured values. Complete only the pending verification items, update the README and checklists from the generated artifact, commit locally, and leave the worktree clean. Do not push.
