package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/madhav-murali/medjed/internal"
	"github.com/madhav-murali/medjed/middleware"
)

func main() {
	limiter := internal.NewSlidingWindowLimiter("localhost:6379", 100, 1*time.Minute)

	app := gin.Default()
	app.Use(middleware.RateLimitMiddleware(limiter))

	app.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	app.Run(":8080")
}
