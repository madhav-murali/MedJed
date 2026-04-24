package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madhav-murali/medjed"
)

// RateLimitMiddleware returns a Gin middleware that enforces rate limiting.
// Pass any implementation of medjed.Limiter.
func RateLimitMiddleware(limiter medjed.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIp := c.ClientIP()
		if limiter.Allow(c, userIp) == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
