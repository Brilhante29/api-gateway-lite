param(
  [switch]$SkipDocker
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$failures = [System.Collections.Generic.List[string]]::new()

function Add-Failure([string]$Message) {
  $script:failures.Add($Message)
}

function Require-File([string]$RelativePath) {
  if (-not (Test-Path -LiteralPath (Join-Path $root $RelativePath) -PathType Leaf)) {
    Add-Failure "Missing file: $RelativePath"
  }
}

function Invoke-Checked([string]$Label, [scriptblock]$Command) {
  & $Command
  if ($LASTEXITCODE -ne 0) {
    Add-Failure "$Label failed with exit code $LASTEXITCODE"
  }
  $global:LASTEXITCODE = 0
}

$requiredFiles = @(
  "README.md",
  "project.yaml",
  "REFERENCES.md",
  "AGENTS.md",
  "compose.yaml",
  "Dockerfile",
  ".dockerignore",
  ".github/workflows/ci.yml",
  "deploy/otel-collector.yaml",
  "tools/run-benchmark.ps1",
  "tools/run-benchmark.sh",
  "sdd/spec.md",
  "sdd/benchmark-plan.md",
  "sdd/architecture-decision.md",
  "sdd/technical-decision.md",
  "sdd/agent-handoff.md",
  "sdd/reuse-improvement-review.md",
  "sdd/release-checklist.md",
  "benchmarks/results/latest.json"
)
foreach ($file in $requiredFiles) { Require-File $file }

$readmePath = Join-Path $root "README.md"
if (Test-Path -LiteralPath $readmePath) {
  $readme = Get-Content -Raw -LiteralPath $readmePath
  if ($readme -notmatch '^# #18 API Gateway Lite') { Add-Failure "README must open with project number and name" }
  if ($readme -match 'Measured result:\*\* pending') { Add-Failure "README still has a pending benchmark result" }
  foreach ($metric in @("overhead_p50_ms", "overhead_p95_ms", "overhead_p99_ms", "gateway_throughput_rps")) {
    if ($readme -notmatch [regex]::Escape($metric)) { Add-Failure "README is missing metric $metric" }
  }
}

$projectPath = Join-Path $root "project.yaml"
if (Test-Path -LiteralPath $projectPath) {
  $project = Get-Content -Raw -LiteralPath $projectPath
  if ($project -notmatch '(?m)^status: benchmarked\r?$') { Add-Failure "project.yaml status must be benchmarked" }
  if ($project -notmatch '(?m)^  primary_metric: overhead_p95_ms\r?$') { Add-Failure "project.yaml primary metric must be overhead_p95_ms" }
}

$reuseReviewPath = Join-Path $root "sdd/reuse-improvement-review.md"
if (Test-Path -LiteralPath $reuseReviewPath) {
  $reuseReview = Get-Content -Raw -LiteralPath $reuseReviewPath
  foreach ($pattern in @(
    '(?m)^- \[x\] Reusable improvements were patched or recorded\.\r?$',
    '(?m)^- \[x\] Project-specific implementation was not moved into the kit\.\r?$',
    '(?m)^- \[x\] Validation reflects .+\.\r?$'
  )) {
    if ($reuseReview -notmatch $pattern) { Add-Failure "Reuse improvement final gate is incomplete" }
  }
  if ($reuseReview -match '<id>|<project-name>|\|  \|') { Add-Failure "Reuse improvement review contains placeholders" }
}

$resultPath = Join-Path $root "benchmarks/results/latest.json"
if (Test-Path -LiteralPath $resultPath) {
  $trackedResult = git -C $root ls-files --error-unmatch benchmarks/results/latest.json 2>$null
  if ($LASTEXITCODE -ne 0 -or -not $trackedResult) { Add-Failure "Benchmark latest.json must be tracked by Git" }
  $global:LASTEXITCODE = 0
  try {
    $artifact = Get-Content -Raw -LiteralPath $resultPath | ConvertFrom-Json
    if ($artifact.schema_version -ne 2) { Add-Failure "Benchmark schema_version must be 2" }
    if ($artifact.project -ne "api-gateway-lite") { Add-Failure "Benchmark project identity is invalid" }
    if ($artifact.benchmark_id -ne "gateway-overhead-v2") { Add-Failure "Benchmark ID is invalid" }
    if ($artifact.execution.repeat -lt 3) { Add-Failure "Benchmark needs at least 3 repetitions" }
    if ($artifact.provenance.source_commit -notmatch '^[0-9a-f]{40}$') { Add-Failure "Benchmark source commit is invalid" }
    if ($artifact.provenance.clean_tree -ne $true) { Add-Failure "Benchmark clean_tree must be true" }
    foreach ($digestName in @("image_digest", "dependency_lock_digest", "artifact_digest")) {
      if ($artifact.provenance.$digestName -notmatch '^sha256:[0-9a-f]{64}$') {
        Add-Failure "Benchmark $digestName is invalid"
      }
    }
    $metricNames = @($artifact.metrics | ForEach-Object { $_.name })
    foreach ($metricName in @(
      "direct_p50_ms", "direct_p95_ms", "direct_p99_ms",
      "gateway_p50_ms", "gateway_p95_ms", "gateway_p99_ms",
      "overhead_p50_ms", "overhead_p95_ms", "overhead_p99_ms",
      "direct_throughput_rps", "gateway_throughput_rps",
      "gateway_rejects", "direct_failures", "gateway_failures"
    )) {
      if ($metricNames -notcontains $metricName) { Add-Failure "Benchmark is missing metric $metricName" }
    }
    foreach ($metric in @($artifact.metrics)) {
      if (@($metric.samples).Count -lt 3) { Add-Failure "Metric $($metric.name) has fewer than 3 samples" }
    }
  } catch {
    Add-Failure "Benchmark JSON could not be parsed: $($_.Exception.Message)"
  }
}

$legacy = ("ro" + "che" + "do")
$patterns = @($legacy, ($legacy.Substring(0, 1).ToUpper() + $legacy.Substring(1)))
$searchFiles = Get-ChildItem -Path $root -Recurse -File | Where-Object {
  $normalized = $_.FullName -replace "\\", "/"
  $normalized -notmatch "/.git/" -and
  $_.Extension -in @(".md", ".yaml", ".yml", ".json", ".ps1", ".sh", ".go")
}
$forbidden = Select-String -Path $searchFiles.FullName -Pattern $patterns -SimpleMatch -ErrorAction SilentlyContinue
if ($forbidden) { Add-Failure "Forbidden legacy nickname or mojibake found" }

$staleDecisionFiles = @(
  (Join-Path $root "project.yaml"),
  (Join-Path $root "README.md"),
  (Join-Path $root "sdd/architecture-decision.md"),
  (Join-Path $root "sdd/technical-decision.md"),
  (Join-Path $root "sdd/agent-handoff.md")
)
$stale = Select-String -Path $staleDecisionFiles -Pattern @("in-memory token bucket", "noop exporter", "Go 1.22") -SimpleMatch -ErrorAction SilentlyContinue
if ($stale) { Add-Failure "Documentation contains superseded runtime decisions" }

Push-Location -LiteralPath $root
try {
  Invoke-Checked "Compose configuration" { docker compose config --quiet }
  if (-not $SkipDocker) {
    $imageName = "api-gateway-lite-local"
    $testImage = "api-gateway-lite-test-local"
    Invoke-Checked "Docker test-stage build" { docker build --target build -t $testImage . }
    Invoke-Checked "Docker runtime build" { docker build -t $imageName . }

    $previousRate = $env:RATE_LIMIT
    $previousBurst = $env:RATE_BURST
    $previousAPIKey = $env:API_KEY
    $env:RATE_LIMIT = "0.01"
    $env:RATE_BURST = "2"
    $env:API_KEY = "local-demo-key"
    try {
      docker compose up --build --wait redis upstream otel-collector gateway
      $composeStarted = $LASTEXITCODE -eq 0
      if (-not $composeStarted) { Add-Failure "Compose startup failed with exit code $LASTEXITCODE" }
      $global:LASTEXITCODE = 0
      if ($composeStarted) {
        Invoke-Checked "Redis integration tests" { docker run --rm --network api-gateway-lite_default -e TEST_REDIS_ADDR=redis:6379 $testImage go test -count=1 ./internal/ratelimit -run RedisLimiter }
        $unauthorizedStatus = curl.exe --silent --output NUL --write-out "%{http_code}" http://localhost:8080/echo
        if ($unauthorizedStatus -ne "401") { Add-Failure "Unauthenticated request returned $unauthorizedStatus instead of 401" }
        $global:LASTEXITCODE = 0

        $headers = Join-Path ([System.IO.Path]::GetTempPath()) "api-gateway-lite-headers.txt"
        Invoke-Checked "Authenticated smoke" { curl.exe --fail --silent --output NUL --dump-header $headers -H "X-API-Key: local-demo-key" -H "X-Correlation-ID: validation-1" http://localhost:8080/echo }
        $headerText = Get-Content -Raw -LiteralPath $headers
        foreach ($pattern in @("X-Correlation-Id: validation-1", "X-Received-Correlation-Id: validation-1", "X-Received-Traceparent: 00-")) {
          if ($headerText -notmatch [regex]::Escape($pattern)) { Add-Failure "Compose smoke is missing header evidence: $pattern" }
        }
        Invoke-Checked "Second quota token" { curl.exe --fail --silent --output NUL -H "X-API-Key: local-demo-key" http://localhost:8080/echo }
        curl.exe --silent --output NUL --write-out "%{http_code}" -H "X-API-Key: local-demo-key" http://localhost:8080/echo | ForEach-Object {
          if ($_ -ne "429") { Add-Failure "Third protected request returned $_ instead of 429" }
        }
        $global:LASTEXITCODE = 0
      }
    } finally {
      $env:RATE_LIMIT = $previousRate
      $env:RATE_BURST = $previousBurst
      $env:API_KEY = $previousAPIKey
      try { docker compose --profile benchmark down --volumes --remove-orphans *> $null } catch { }
    }
  }
} finally {
  Pop-Location
}

if ($failures.Count -gt 0) {
  Write-Host "portfolio project validation failed with $($failures.Count) issue(s):"
  foreach ($failure in $failures) {
    Write-Host "  - $failure"
    if ($env:GITHUB_ACTIONS -eq "true") { Write-Host "::error::$failure" }
  }
  exit 1
}

Write-Host "portfolio project validation passed"
exit 0
