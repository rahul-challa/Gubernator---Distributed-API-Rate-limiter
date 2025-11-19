package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gubernator/pkg/limiter"
)

// Example usage of the rate limiter
func main() {
	// Create rate limiter configuration
	config := limiter.Config{
		Capacity:  10,
		RefillRate: 1.0,
		RedisAddr: "localhost:6379",
		RedisDB:   0,
	}

	// Initialize rate limiter
	rl, err := limiter.NewRateLimiter(config)
	if err != nil {
		log.Fatalf("Failed to create rate limiter: %v", err)
	}
	defer rl.Close()

	ctx := context.Background()
	key := "ratelimit:example-user-123"

	// Simulate multiple requests
	for i := 0; i < 15; i++ {
		result, err := rl.Allow(ctx, key)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		if result.Allowed {
			fmt.Printf("Request %d: [ALLOWED] Remaining: %d, Reset: %s\n",
				i+1, result.Remaining, result.ResetTime.Format(time.RFC3339))
		} else {
			fmt.Printf("Request %d: [BLOCKED] Remaining: %d, Reset: %s\n",
				i+1, result.Remaining, result.ResetTime.Format(time.RFC3339))
		}

		time.Sleep(200 * time.Millisecond)
	}
}

