package redisThrottler

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

/*
TestThrottle requires local Redis instance to be running on localhost:6379.
This is not a typical unit test, but rather a demonstration of the functionality.
*/
func TestThrottle(t *testing.T) {
	// Create a context for the test
	ctx := context.Background()

	// Create a new Redis client for testing
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	//defer redisClient.Close()
	//
	// Create a new RateLimitThrottler instance
	throttler := &RateLimitThrottler{
		redisClient:     redisClient,
		rateLimitWindow: 5 * time.Second,
		requestLimit:    5,
	}

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
		panic(err)
	}

	// Test case: Requests within the limit should not be throttled
	uniqueToken := "abc"
	for i := 0; i < throttler.requestLimit; i++ {
		throttled := throttler.Throttle(ctx, uniqueToken)
		assert.False(t, throttled, "Request should not be throttled")
	}

	// Test case: Requests exceeding the limit should be throttled
	throttled := throttler.Throttle(ctx, uniqueToken)
	assert.True(t, throttled, "Request should be throttled")

	// Test case: After the rate limit window expires, requests should not be throttled
	time.Sleep(throttler.rateLimitWindow + 1*time.Second)
	throttled = throttler.Throttle(ctx, uniqueToken)
	assert.False(t, throttled, "Request should not be throttled after rate limit window expires")
}
