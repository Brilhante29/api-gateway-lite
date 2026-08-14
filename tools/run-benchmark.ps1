param(
  [string]$ResultsDir = ""
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Push-Location -LiteralPath $root
try {
  $dirty = git status --porcelain
  if ($dirty) {
    throw "Benchmark provenance requires a clean Git worktree. Commit the implementation first."
  }

  $sourceCommit = (git rev-parse HEAD).Trim()
  $shortCommit = $sourceCommit.Substring(0, 12)
  $imageRef = "api-gateway-lite:benchmark-$shortCommit"
  $dependencyDigest = "sha256:" + (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $root "go.sum")).Hash.ToLowerInvariant()

  docker build --tag $imageRef .
  if ($LASTEXITCODE -ne 0) { throw "docker build failed" }
  $imageDigest = (docker image inspect $imageRef --format "{{.Id}}").Trim()
  if ($LASTEXITCODE -ne 0 -or $imageDigest -notmatch '^sha256:[0-9a-f]{64}$') {
    throw "could not resolve the benchmark image digest"
  }

  $env:SOURCE_COMMIT = $sourceCommit
  $env:IMAGE_REF = $imageRef
  $env:IMAGE_DIGEST = $imageDigest
  $env:DEPENDENCY_LOCK_DIGEST = $dependencyDigest
  $env:BENCHMARK_CLEAN_TREE = "true"
  $env:BENCHMARK_COMMAND = "pwsh ./tools/run-benchmark.ps1"
  if (-not $env:BENCHMARK_PRODUCER) { $env:BENCHMARK_PRODUCER = "local" }
  if (-not $env:BENCHMARK_HARDWARE_CLASS) { $env:BENCHMARK_HARDWARE_CLASS = "docker-desktop" }
  if ([string]::IsNullOrWhiteSpace($ResultsDir)) {
    $ResultsDir = Join-Path $root "benchmarks/results"
  } elseif (-not [System.IO.Path]::IsPathRooted($ResultsDir)) {
    $ResultsDir = Join-Path $root $ResultsDir
  }
  New-Item -ItemType Directory -Force -Path $ResultsDir | Out-Null
  $env:BENCHMARK_RESULTS_DIR = (Resolve-Path -LiteralPath $ResultsDir).Path

  docker compose --profile benchmark up --no-build --abort-on-container-exit --exit-code-from benchmark benchmark
  if ($LASTEXITCODE -ne 0) { throw "benchmark compose run failed" }
  if (-not (Test-Path -LiteralPath (Join-Path $env:BENCHMARK_RESULTS_DIR "latest.json") -PathType Leaf)) {
    throw "benchmark did not write latest.json to the selected results directory"
  }
} finally {
  try { docker compose --profile benchmark down --volumes --remove-orphans *> $null } catch { }
  Pop-Location
}
