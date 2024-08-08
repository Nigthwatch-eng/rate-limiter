package main

import (
	redisThrottler "RateLimiter/redis-throttler"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()
var rateLimiterClient redisThrottler.RateLimiter

func main() {
	logrus.Info("init server with rate limiter...")

	rateLimiterClient = redisThrottler.NewRedisClient(ctx)

	r := gin.Default()

	r.POST("/api/configure", updateRateLimiterConfig)

	// User middleware for rate limiter check
	r.GET("/api/is_rate_limited/:unique_token", rateLimitHandler)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logrus.Info("starting server with rate limiter...")
	_ = s.ListenAndServe()
}

// rateLimitHandler is a handler function for rate limiting requests
// It takes a *gin.Context as an argument, which carries the request details and decorates rate_limited response with true or false
func rateLimitHandler(c *gin.Context) {
	uniqueToken := c.Param("unique_token")
	if uniqueToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unique_token is required"})
		return
	}

	rateLimited := rateLimiterClient.Throttle(ctx, uniqueToken)
	c.JSON(http.StatusOK, gin.H{"rate_limited": rateLimited})
}

// updateRateLimiterConfig is a handler function for updating the rate limiter configuration
// Currently this method is not implemented
func updateRateLimiterConfig(c *gin.Context) {
	logrus.Warn("API not implemented")
}
