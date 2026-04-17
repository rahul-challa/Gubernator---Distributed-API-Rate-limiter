package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds the configuration for the rate limiter
type Config struct {
	Capacity   int     // Maximum number of tokens in the bucket
	RefillRate float64 // Tokens per second
	RedisAddr  string  // Redis server address
	RedisDB    int     // Redis database number
}

// RateLimiter implements the token bucket algorithm using Redis
type RateLimiter struct {
	client     *redis.Client
	capacity   int
	refillRate float64
	luaScript  *redis.Script
}

// Result represents the result of a rate limit check
type Result struct {
	Allowed   bool
	Remaining int
	ResetTime time.Time
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(config Config) (*RateLimiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
		DB:   config.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Load Lua script
	luaScript := redis.NewScript(luaScriptSource)

	return &RateLimiter{
		client:     client,
		capacity:   config.Capacity,
		refillRate: config.RefillRate,
		luaScript:  luaScript,
	}, nil
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	now := time.Now().Unix()

	// Execute Lua script atomically
	result, err := rl.luaScript.Run(ctx, rl.client, []string{key},
		rl.capacity,
		rl.refillRate,
		now).Result()

	if err != nil {
		return nil, fmt.Errorf("lua script execution failed: %w", err)
	}

	// Parse result: [allowed, remaining, reset_time]
	res := result.([]interface{})
	allowed := res[0].(int64) == 1
	remaining := int(res[1].(int64))
	resetTime := time.Unix(res[2].(int64), 0)

	return &Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: resetTime,
	}, nil
}

// Close closes the Redis connection
func (rl *RateLimiter) Close() error {
	return rl.client.Close()
}

// Lua script source (embedded)
const luaScriptSource = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')

local tokens = tonumber(bucket[1]) or capacity
local last_refill = tonumber(bucket[2]) or now

local time_passed = now - last_refill

if time_passed > 0 then
    tokens = math.min(capacity, tokens + (refill_rate * time_passed))
    last_refill = now
end

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call('HMSET', key, 'tokens', tokens, 'last_refill', last_refill)
redis.call('EXPIRE', key, 3600)

-- Calculate reset time: when bucket will be full again
local reset_time = last_refill
if tokens < capacity then
    local tokens_needed = capacity - tokens
    reset_time = last_refill + math.ceil(tokens_needed / refill_rate)
end

return {allowed, tokens, reset_time}
`
