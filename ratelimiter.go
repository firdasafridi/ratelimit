package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

// RateLimiter struct holds the configurations and states for the rate limiter.
type RateLimiter struct {
	redisClient *redis.Client              // Redis client used for storing request counts.
	limiterMap  map[string]*rate.Limiter   // In-memory map to hold the rate limiters for different users and endpoints.
	mu          sync.Mutex                 // Mutex to ensure thread safety when accessing the limiterMap.
	defaultConf RateLimitConfig            // Default rate limiting configuration.
	endpointMap map[string]RateLimitConfig // Map to store custom rate limiting configurations for specific endpoints.
}

// RateLimitConfig struct holds the configuration for rate limiting.
type RateLimitConfig struct {
	MaxRequests int           // Maximum number of requests allowed in a given interval.
	Interval    time.Duration // The interval in which MaxRequests are counted.
}

// NewRateLimiter function initializes and returns a new RateLimiter instance.
// It accepts a Redis client and a default rate limiting configuration.
func NewRateLimiter(rdb *redis.Client, defaultConf RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redisClient: rdb,
		limiterMap:  make(map[string]*rate.Limiter),
		defaultConf: defaultConf,
		endpointMap: make(map[string]RateLimitConfig),
	}
}

// AddCustomRateLimit function adds a custom rate limiting configuration for a specific endpoint.
// It is thread-safe and can be called concurrently.
func (rl *RateLimiter) AddCustomRateLimit(endpoint string, conf RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.endpointMap[endpoint] = conf
}

// getRateLimitConfig function retrieves the rate limiting configuration for a given endpoint.
// If there's no custom configuration for the endpoint, it returns the default configuration.
// It is thread-safe and can be called concurrently.
func (rl *RateLimiter) getRateLimitConfig(endpoint string) RateLimitConfig {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if conf, exists := rl.endpointMap[endpoint]; exists {
		return conf
	}
	return rl.defaultConf
}

// Allow function checks if a request from a user to a specific endpoint should be allowed or denied based on the rate limiting rules.
// It is not fully thread-safe as of this implementation and should not be called concurrently without consideration.
func (rl *RateLimiter) Allow(ctx context.Context, userId, endpoint string) bool {
	key := fmt.Sprintf("%s:%s", userId, endpoint)
	limiter, exists := rl.limiterMap[key]
	conf := rl.getRateLimitConfig(endpoint)

	if !exists {
		limiter = rate.NewLimiter(rate.Every(conf.Interval), conf.MaxRequests)
		rl.limiterMap[key] = limiter
	}

	count, err := rl.redisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false
	}

	if count < conf.MaxRequests {
		rl.redisClient.Set(ctx, key, count+1, conf.Interval)
		return true
	}

	return limiter.Allow()
}
