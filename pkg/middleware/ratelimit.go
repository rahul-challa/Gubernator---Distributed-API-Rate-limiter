package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gubernator/pkg/limiter"
)

// RateLimitMiddleware creates HTTP middleware for rate limiting
func RateLimitMiddleware(limiter *limiter.RateLimiter, keyExtractor KeyExtractor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract key (IP address or API key)
			key := keyExtractor(r)

			// Check rate limit
			result, err := limiter.Allow(r.Context(), key)
			if err != nil {
				http.Error(w, "Rate limiter error", http.StatusInternalServerError)
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", "10")
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", result.ResetTime.Format(http.TimeFormat))

			if !result.Allowed {
				// Return 429 Too Many Requests
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				retryAfter := int(result.ResetTime.Sub(time.Now()).Seconds())
				if retryAfter < 0 {
					retryAfter = 0
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":       "Rate limit exceeded",
					"retry_after": retryAfter,
				})
				return
			}

			// Request allowed, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// KeyExtractor function type for extracting rate limit keys
type KeyExtractor func(*http.Request) string

// ExtractIP extracts the client IP address from the request
func ExtractIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// ExtractAPIKey extracts API key from Authorization header or query parameter
func ExtractAPIKey(r *http.Request) string {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if strings.HasPrefix(auth, "ApiKey ") {
		return strings.TrimPrefix(auth, "ApiKey ")
	}

	// Check query parameter
	apiKey := r.URL.Query().Get("api_key")
	if apiKey != "" {
		return apiKey
	}

	// Fall back to IP if no API key found
	return ExtractIP(r)
}

// DefaultKeyExtractor uses IP address as the default key
func DefaultKeyExtractor(r *http.Request) string {
	return "ratelimit:" + ExtractIP(r)
}
