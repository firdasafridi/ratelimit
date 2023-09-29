# Rate Limiter

A Go library for rate limiting user requests using both in-memory and Redis storage to ensure optimal performance and scalability.

## Features
- In-memory rate limiting using token bucket algorithm.
- Redis backed storage for distributed rate limiting across multiple instances.
- Supports custom rate limiting rules per endpoint.
- Thread-safe operations for adding and retrieving configurations.

## Installation

To install the `ratelimiter` package, run the following command:

```sh
go get github.com/firdasafridi/ratelimiter
```

## Usage
Here's a basic example of how to use the ratelimiter library.

## Creating a New Rate Limiter
```go
package main

import (
	"time"
	"github.com/go-redis/redis/v8"
	"github.com/firdasafridi/ratelimiter"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0, // use default DB
	})
	
	defaultConf := ratelimiter.RateLimitConfig{
		MaxRequests: 100,
		Interval:    time.Minute,
	}

	rl := ratelimiter.NewRateLimiter(rdb, defaultConf)
}

```
## Adding a Custom Rate Limit for an Endpoint
```go
rl.AddCustomRateLimit("/api/sensitive-endpoint", ratelimiter.RateLimitConfig{
	MaxRequests: 10,
	Interval:    time.Minute,
})
```
Checking if a Request is Allowed
```go
import "context"

isAllowed := rl.Allow(context.TODO(), "userID", "/api/sensitive-endpoint")
if isAllowed {
	// Process the request
} else {
	// Deny the request
}
```


## Configuration
The rate limiter uses the following configurations:

`MaxRequests`: The maximum number of requests allowed in a specified interval.

`Interval`: The time duration in which the MaxRequests are counted.

These configurations can be set globally for all endpoints and can also be customized per endpoint.
