package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/firdasafridi/ratelimiter"
	"github.com/go-redis/redis/v8"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPass := os.Getenv("REDIS_PASS")

	if redisAddr == "" {
		log.Fatal("Environment variable REDIS_ADDR is required")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass, // no password set
		DB:       0,         // use default DB
	})

	defaultConf := ratelimiter.RateLimitConfig{
		MaxRequests: 5,
		Interval:    10 * time.Second,
	}

	limiter := ratelimiter.NewRateLimiter(rdb, defaultConf)

	customConf := ratelimiter.RateLimitConfig{
		MaxRequests: 2,
		Interval:    5 * time.Second,
	}

	limiter.AddCustomRateLimit("/v1/special", customConf)

	userId := "user1"
	endpoints := []string{"/v1/ping", "/v1/special"}
	ctx := context.TODO()

	for _, endpoint := range endpoints {
		fmt.Printf("Testing endpoint: %s\n", endpoint)
		for i := 0; i < 100; i++ {
			if limiter.Allow(ctx, userId, endpoint) {
				fmt.Printf("Request %d to %s allowed\n", i+1, endpoint)
			} else {
				fmt.Printf("Request %d to %s denied - too many requests\n", i+1, endpoint)
				break
			}
		}
		fmt.Println("---------------------------")
	}
}
