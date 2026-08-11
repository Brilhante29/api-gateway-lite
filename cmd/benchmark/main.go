package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Brilhante29/api-gateway-lite/internal/benchmark"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := benchmark.Run(ctx, benchmark.Config{
		GatewayURL:           env("BENCHMARK_GATEWAY_URL", "http://gateway-benchmark:8080"),
		DirectURL:            env("BENCHMARK_DIRECT_URL", "http://upstream:8081"),
		APIKey:               env("API_KEY", "benchmark-key"),
		WarmupIterations:     envInt("BENCHMARK_WARMUP", 100),
		MeasuredIterations:   envInt("BENCHMARK_REQUESTS", 1000),
		Concurrency:          envInt("BENCHMARK_CONCURRENCY", 16),
		Repeat:               envInt("BENCHMARK_REPEAT", 3),
		Command:              env("BENCHMARK_COMMAND", "docker compose --profile benchmark up --abort-on-container-exit --exit-code-from benchmark benchmark"),
		SourceCommit:         os.Getenv("SOURCE_COMMIT"),
		CleanTree:            env("BENCHMARK_CLEAN_TREE", "false") == "true",
		ImageRef:             os.Getenv("IMAGE_REF"),
		ImageDigest:          os.Getenv("IMAGE_DIGEST"),
		DependencyLockDigest: os.Getenv("DEPENDENCY_LOCK_DIGEST"),
		Producer:             env("BENCHMARK_PRODUCER", "local"),
		CIRunURL:             os.Getenv("CI_RUN_URL"),
		HardwareClass:        env("BENCHMARK_HARDWARE_CLASS", "docker-desktop"),
	})
	if err != nil {
		log.Fatalf("benchmark: %v", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("encode benchmark: %v", err)
	}
	data = append(data, '\n')
	resultPath := env("BENCHMARK_RESULT_PATH", "/results/latest.json")
	if err := os.WriteFile(resultPath, data, 0o644); err != nil {
		log.Fatalf("write benchmark result: %v", err)
	}
	fmt.Print(string(data))
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("%s must be an integer: %v", name, err)
	}
	return parsed
}
