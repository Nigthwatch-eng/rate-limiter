package redisThrottler

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RateLimiter interface {
	// Throttle returns true when limit is exceeded, and false otherwise. key is a unique token
	Throttle(ctx context.Context, key string) bool
}

type RateLimitThrottler struct {
	redisClient     *redis.Client
	rateLimitWindow time.Duration
	requestLimit    int
}

func NewRedisClient(ctx context.Context) RateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
		panic(err)
	}

	return &RateLimitThrottler{
		redisClient:     client,
		rateLimitWindow: 30 * time.Second, // default interval
		requestLimit:    2,                // default limit
	}
}

/*
Throttle is a method that checks if a request should be throttled based on the rate limiting rules.
It takes a context and a unique token as input, and returns a boolean indicating whether the request should be throttled(rate limited) or not.
Lua script is used to achieve atomicity for transactions by allowing redis to execute script on the cluster.
Refer to https://redis.io/docs/latest/develop/interact/programmability/eval-intro/ for more information on lua script.
*/
func (r RateLimitThrottler) Throttle(ctx context.Context, token string) bool {
	key := r.getSlidingWindowKey(token)
	const script = `
					local key=KEYS[1]
					local ttl=ARGV[1]
					local nowTime=ARGV[2]
					local expirationTime=ARGV[3]
					local maxElements=tonumber(ARGV[4])
					local readOnly=tonumber(ARGV[5])
					-- delete all entries that have expired
					redis.call('ZREMRANGEBYSCORE', key, '-inf', nowTime)
					local card = tonumber(redis.call('ZCARD', key))
					-- If there is space in the window
					if maxElements >= card+1 then
						if readOnly == 0 then
							-- Add the entry, use the score as the name
							redis.call('ZADD', key, expirationTime, expirationTime)
							redis.call('EXPIRE', key, ttl)
						end
						return 1
					end
					return 0
					`

	now := time.Now()
	t := int(r.rateLimitWindow.Seconds())
	nowTime := now.UnixNano() / int64(time.Microsecond)
	expirationTime := now.Add(r.rateLimitWindow).UnixNano() / int64(time.Microsecond)

	cmd := r.redisClient.Eval(ctx, script, []string{key}, t, nowTime, expirationTime, r.requestLimit, false)
	if cmd.Err() != nil {
		panic(cmd.Err())
	}

	i, err := cmd.Result()
	if err != nil {
		// This should never happen
		panic("internal error converting output value")
	}
	return i.(int64) < 1
}

/*
ThrottleNonAtomic is a method that checks if a request should be throttled based on the rate limiting rules.
It takes a context and a unique token as input, and returns a boolean indicating whether the request should be throttled(rate limited) or not.
This is method does not provide atomicity and hence it is not used in the code path.
Kept it here as it was my initial implementation prior to using the Lua script.
Deprecated: ThrottleNonAtomic is deprecated, do not use. 
*/
func (r RateLimitThrottler) ThrottleNonAtomic(ctx context.Context, token string) bool {
	key := r.getSlidingWindowKey(token)
	now := time.Now().Unix()
	windowStart := now - int64(r.rateLimitWindow.Seconds())

	// Remove timestamps outside the current window
	r.redisClient.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// Count the number of requests in the current window
	reqCount, err := r.redisClient.ZCount(ctx, key, strconv.FormatInt(windowStart, 10), strconv.FormatInt(now, 10)).Result()
	if err != nil {
		panic(err)
	}

	// Check if the rate limit is exceeded
	if reqCount >= int64(r.requestLimit) {
		return true
	}

	// Add the current request timestamp
	r.redisClient.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: now})
	// Set an expiry for the key slightly longer than the window
	r.redisClient.Expire(ctx, key, r.rateLimitWindow+time.Second)

	return false
}

// getSlidingWindowKey prefix the Redis key, as the cluster may be using a shared Redis instance,
// and we don't want to collide with other instances.
// The key is in the format "rate_limit:{uniqueToken}"
func (r RateLimitThrottler) getSlidingWindowKey(uniqueToken string) string {
	return "rate_limit:" + uniqueToken
}
