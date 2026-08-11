package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	UpstreamURL     string
	ListenAddr      string
	APIKey          string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	RateLimit       float64
	RateBurst       int
	RateLimitPrefix string
	Telemetry       string
	ServiceName     string
}

func LoadConfig() Config {
	return Config{
		UpstreamURL:     getEnv("UPSTREAM_URL", "http://localhost:8081"),
		ListenAddr:      getEnv("LISTEN_ADDR", ":8080"),
		APIKey:          getEnv("API_KEY", "local-demo-key"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		RedisDB:         getEnvInt("REDIS_DB", 0),
		RateLimit:       getEnvFloat("RATE_LIMIT", 100),
		RateBurst:       getEnvInt("RATE_BURST", 200),
		RateLimitPrefix: getEnv("RATE_LIMIT_PREFIX", "api-gateway-lite:ratelimit"),
		Telemetry:       strings.ToLower(getEnv("TELEMETRY", "otlp")),
		ServiceName:     getEnv("OTEL_SERVICE_NAME", "api-gateway-lite"),
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("API_KEY must not be empty")
	}
	if strings.TrimSpace(c.RedisAddr) == "" {
		return fmt.Errorf("REDIS_ADDR must not be empty")
	}
	if c.RedisDB < 0 {
		return fmt.Errorf("REDIS_DB must be non-negative")
	}
	if c.RateLimit <= 0 {
		return fmt.Errorf("RATE_LIMIT must be greater than zero")
	}
	if c.RateBurst <= 0 {
		return fmt.Errorf("RATE_BURST must be greater than zero")
	}
	if strings.TrimSpace(c.RateLimitPrefix) == "" {
		return fmt.Errorf("RATE_LIMIT_PREFIX must not be empty")
	}
	if c.Telemetry != "otlp" && c.Telemetry != "none" {
		return fmt.Errorf("TELEMETRY must be otlp or none")
	}
	return nil
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
