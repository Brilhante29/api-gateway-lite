package ratelimit

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

type Decision struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAfter time.Duration
	RetryAfter time.Duration
}

type Limiter interface {
	Allow(ctx context.Context, subject string) (Decision, error)
}

type RedisLimiter struct {
	client    *redis.Client
	rate      float64
	burst     int
	keyPrefix string
}

// Redis TIME and one Lua script make refill and consumption atomic across all
// gateway instances without trusting their local clocks.
var tokenBucketScript = redis.NewScript(`
local current = redis.call('TIME')
local now_ms = (tonumber(current[1]) * 1000) + math.floor(tonumber(current[2]) / 1000)
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])

local bucket = redis.call('HMGET', KEYS[1], 'tokens', 'updated_at_ms')
local tokens = tonumber(bucket[1])
local updated_at_ms = tonumber(bucket[2])

if tokens == nil then
  tokens = capacity
  updated_at_ms = now_ms
else
  local elapsed_seconds = math.max(0, now_ms - updated_at_ms) / 1000
  tokens = math.min(capacity, tokens + (elapsed_seconds * rate))
end

local allowed = 0
if tokens >= 1 then
  allowed = 1
  tokens = tokens - 1
end

local remaining = math.floor(tokens)
local reset_ms = math.ceil(((capacity - tokens) / rate) * 1000)
local retry_ms = 0
if allowed == 0 then
  retry_ms = math.ceil(((1 - tokens) / rate) * 1000)
end

redis.call('HSET', KEYS[1], 'tokens', tokens, 'updated_at_ms', now_ms)
redis.call('PEXPIRE', KEYS[1], math.max(1000, math.ceil((capacity / rate) * 2000)))

return {allowed, remaining, reset_ms, retry_ms}
`)

func NewRedisLimiter(client *redis.Client, rate float64, burst int, keyPrefix string) *RedisLimiter {
	return &RedisLimiter{
		client:    client,
		rate:      rate,
		burst:     burst,
		keyPrefix: keyPrefix,
	}
}

func NewRedisClient(addr, password string, database int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           database,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolTimeout:  2 * time.Second,
	})
}

func (l *RedisLimiter) Allow(ctx context.Context, subject string) (Decision, error) {
	if subject == "" {
		return Decision{}, fmt.Errorf("rate-limit subject must not be empty")
	}

	values, err := tokenBucketScript.Run(
		ctx,
		l.client,
		[]string{fmt.Sprintf("%s:{%s}", l.keyPrefix, subject)},
		l.rate,
		l.burst,
	).Int64Slice()
	if err != nil {
		return Decision{}, fmt.Errorf("run Redis token bucket: %w", err)
	}
	if len(values) != 4 {
		return Decision{}, fmt.Errorf("Redis token bucket returned %d values, want 4", len(values))
	}

	return Decision{
		Allowed:    values[0] == 1,
		Limit:      l.burst,
		Remaining:  int(values[1]),
		ResetAfter: time.Duration(values[2]) * time.Millisecond,
		RetryAfter: time.Duration(values[3]) * time.Millisecond,
	}, nil
}

func (l *RedisLimiter) Ping(ctx context.Context) error {
	return l.client.Ping(ctx).Err()
}

func (l *RedisLimiter) Close() error {
	return l.client.Close()
}

func secondsCeil(duration time.Duration) int64 {
	return int64(math.Max(1, math.Ceil(duration.Seconds())))
}
