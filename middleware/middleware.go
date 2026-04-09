package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madhav-murali/medjed/internal"
)

func RateLimitMiddleware(limiter internal.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIp := c.ClientIP()
		if limiter.Allow(c, userIp) == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "woah slow down buddy ):<"})
			return
		}
		c.Next()
		// lat := time.Since(time.Unix(0, time.Now().UnixNano()))
		// log.Printf("request took: %d ns", lat.Nanoseconds())
	}
}
