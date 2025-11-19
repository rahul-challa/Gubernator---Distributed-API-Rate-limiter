-- Token Bucket Rate Limiting Lua Script
-- This script atomically checks and updates the token bucket for a given key
-- 
-- KEYS[1]: The rate limit key (e.g., "ratelimit:192.168.1.1")
-- ARGV[1]: Bucket capacity (max tokens)
-- ARGV[2]: Refill rate (tokens per second)
-- ARGV[3]: Current timestamp in seconds

local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- Get current bucket state
local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')

local tokens = tonumber(bucket[1]) or capacity
local last_refill = tonumber(bucket[2]) or now

-- Calculate time passed since last refill
local time_passed = now - last_refill

-- Refill tokens based on time passed
if time_passed > 0 then
    tokens = math.min(capacity, tokens + (refill_rate * time_passed))
    last_refill = now
end

-- Check if request is allowed
local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

-- Update bucket state atomically
redis.call('HMSET', key, 'tokens', tokens, 'last_refill', last_refill)
redis.call('EXPIRE', key, 3600) -- Expire after 1 hour of inactivity

-- Calculate reset time: when bucket will be full again
local reset_time = last_refill
if tokens < capacity then
    local tokens_needed = capacity - tokens
    reset_time = last_refill + math.ceil(tokens_needed / refill_rate)
end

-- Return: allowed (1 or 0), remaining tokens, reset time
return {allowed, tokens, reset_time}

