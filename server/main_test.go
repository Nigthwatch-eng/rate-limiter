package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockRateLimiter is a mock implementation of the RateLimiter interface
type mockRateLimiter struct {
	throttleResult bool
}

// Throttle is a mock implementation of the Throttle method
func (m *mockRateLimiter) Throttle(ctx context.Context, uniqueToken string) bool {
	return m.throttleResult
}

func TestRateLimitHandler(t *testing.T) {
	// Create a new instance of the Gin router
	r := gin.Default()

	// Create a mock RateLimiter instance
	mockRL := &mockRateLimiter{throttleResult: false}
	rateLimiterClient = mockRL

	// Define the route for the rateLimitHandler
	r.GET("/api/is_rate_limited/:unique_token", rateLimitHandler)

	// Test case 1: unique_token is provided, and rate limiting is not applied
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/is_rate_limited/test_token", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"rate_limited\":false}", w.Body.String())

	// Test case 2: unique_token is not provided
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/is_rate_limited/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "404 page not found", w.Body.String())

	// Test case 3: unique_token is provided, and rate limiting is applied
	mockRL.throttleResult = true
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/is_rate_limited/test_token", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"rate_limited\":true}", w.Body.String())
}
