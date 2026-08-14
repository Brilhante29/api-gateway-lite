#!/usr/bin/env sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

if [ -n "$(git status --porcelain)" ]; then
  echo "Benchmark provenance requires a clean Git worktree. Commit the implementation first." >&2
  exit 1
fi

SOURCE_COMMIT=$(git rev-parse HEAD)
short_commit=$(printf '%s' "$SOURCE_COMMIT" | cut -c1-12)
IMAGE_REF="api-gateway-lite:benchmark-$short_commit"
DEPENDENCY_LOCK_DIGEST="sha256:$(sha256sum go.sum | cut -d ' ' -f1)"

docker build --tag "$IMAGE_REF" .
IMAGE_DIGEST=$(docker image inspect "$IMAGE_REF" --format '{{.Id}}')

export SOURCE_COMMIT IMAGE_REF IMAGE_DIGEST DEPENDENCY_LOCK_DIGEST
export BENCHMARK_CLEAN_TREE=true
export BENCHMARK_COMMAND="sh ./tools/run-benchmark.sh"
export BENCHMARK_PRODUCER="${BENCHMARK_PRODUCER:-local}"
export BENCHMARK_HARDWARE_CLASS="${BENCHMARK_HARDWARE_CLASS:-docker-linux}"

cleanup() {
  docker compose --profile benchmark down --volumes --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

docker compose --profile benchmark up --no-build --abort-on-container-exit --exit-code-from benchmark benchmark
test -f benchmarks/results/latest.json
