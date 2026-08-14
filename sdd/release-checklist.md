# Release Checklist

- [x] Gateway, upstream, Redis, and OTLP collector have one Docker Compose path.
- [x] Default demo needs no paid credential.
- [x] Unit, real Redis, OTLP export, and Compose smoke tests are defined.
- [x] CI is pinned, least-privilege, bounded, and uploads benchmark evidence.
- [ ] Benchmark V2 was generated from a clean exact commit.
- [ ] README opens with generated p50/p95/p99 overhead and throughput numbers.
- [ ] Project validator passes with Docker enabled.
- [ ] Final local commit exists and worktree is clean.
- [x] Repository remains unpushed as explicitly requested.
