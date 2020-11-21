package gin_redisgo_cooldowns

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func NewRateLimit(rdb *redis.Client, key string, requests int64, time time.Duration, keySuffixGetter func(*gin.Context) string) gin.HandlerFunc {

	if keySuffixGetter == nil {
		keySuffixGetter = KeySuffixGetterIP()
	}

	return func(ctx *gin.Context) {
		// Initialise requestNumber at 0 because they're cool!
		requestNumber := int64(0)

		// Fetch the suffix the the redis key
		// Return if it's empty.
		keySuffix := strings.TrimSpace(keySuffixGetter(ctx))
		if keySuffix == "" {
			return
		}

		// Get the actual number of requests from the database.
		// Also increments. (INCR spits out the old value, and inserts if new)
		res := rdb.Incr(ctx, key+keySuffix)
		if err := res.Err(); res.Err() != nil {
			panic(err)
			return
		} else {
			if requestNumber, err = res.Result(); err != nil {
				panic(err)
			}
		}

		// Check if the current number of requests is greater than allowed.
		if requestNumber >= requests {
			// Reset the expiry to give a fresh cooldown.
			// Set expire once per cool down.
			if requestNumber == requests {
				rdb.Expire(ctx, key+keySuffix, time)
			}

			// Abort and error
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "too many requests",
			})
			return
		} else if requestNumber <= 1 {
			// If it's a new entry in redis, expire the key in a defined time.
			rdb.Expire(ctx, key+keySuffix, time)
		}

		// Give them their request number.
		ctx.Header("X-REQUEST-NUMBER", fmt.Sprint(requestNumber))
		ctx.Next()
	}
}

func KeySuffixGetterIP() func(*gin.Context) string {
	return func(ctx *gin.Context) string {
		return ctx.ClientIP()
	}
}
