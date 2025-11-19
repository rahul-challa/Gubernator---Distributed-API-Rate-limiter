package grpc

import (
	"context"
	"time"

	"gubernator/pkg/limiter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimitService implements the gRPC rate limit service
type RateLimitService struct {
	limiter *limiter.RateLimiter
}

// NewRateLimitService creates a new gRPC rate limit service
func NewRateLimitService(limiter *limiter.RateLimiter) *RateLimitService {
	return &RateLimitService{
		limiter: limiter,
	}
}

// CheckRateLimit checks if a request should be allowed
// This is a placeholder - actual implementation would use generated proto code
func (s *RateLimitService) CheckRateLimit(ctx context.Context, key string) (allowed bool, remaining int, resetTime time.Time, err error) {
	result, err := s.limiter.Allow(ctx, key)
	if err != nil {
		return false, 0, time.Time{}, status.Errorf(codes.Internal, "rate limit check failed: %v", err)
	}

	return result.Allowed, result.Remaining, result.ResetTime, nil
}

