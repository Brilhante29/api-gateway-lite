package internal

import (
	"os"
	"strconv"
)

type Config struct {
	UpstreamURL string
	ListenAddr  string
	APIKey      string
	RateLimit   float64
	RateBurst   int
	Telemetry   string
}

func LoadConfig() Config {
	return Config{
		UpstreamURL: getEnv("UPSTREAM_URL", "http://localhost:8081"),
		ListenAddr:  getEnv("LISTEN_ADDR", ":8080"),
		APIKey:      getEnv("API_KEY", "secret-key-123"),
		RateLimit:   getEnvFloat("RATE_LIMIT", 100),
		RateBurst:   getEnvInt("RATE_BURST", 200),
		Telemetry:   getEnv("TELEMETRY", "none"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
