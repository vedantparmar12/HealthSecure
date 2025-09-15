package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	requests = make(map[string][]int64)
	mu       sync.Mutex
)

func RateLimiter(limit int, per time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		defer mu.Unlock()

		now := time.Now().UnixNano()
		if _, found := requests[ip]; !found {
			requests[ip] = []int64{now}
			c.Next()
			return
		}

		// Remove old timestamps
		var newTimestamps []int64
		for _, ts := range requests[ip] {
			if now-ts < per.Nanoseconds() {
				newTimestamps = append(newTimestamps, ts)
			}
		}
		requests[ip] = newTimestamps

		if len(requests[ip]) >= limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		requests[ip] = append(requests[ip], now)
		c.Next()
	}
}
