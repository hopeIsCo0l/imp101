package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientWindow struct {
	count     int
	windowEnd time.Time
}

var (
	rlMu       sync.Mutex
	rateLimits = map[string]*clientWindow{}
)

func RateLimit(maxPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		rlMu.Lock()
		entry, exists := rateLimits[ip]
		if !exists || now.After(entry.windowEnd) {
			entry = &clientWindow{
				count:     0,
				windowEnd: now.Add(1 * time.Minute),
			}
			rateLimits[ip] = entry
		}
		entry.count++
		currentCount := entry.count
		rlMu.Unlock()

		if currentCount > maxPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests, please try again later"})
			c.Abort()
			return
		}

		c.Next()
	}
}
