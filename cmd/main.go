package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/madhav-murali/medjed"
	"github.com/madhav-murali/medjed/middleware"
)

func main() {
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	port := getEnv("PORT", "8080")
	rateLimit, _ := strconv.ParseInt(getEnv("RATE_LIMIT", "100"), 10, 64)
	rateWindow, _ := time.ParseDuration(getEnv("RATE_WINDOW", "1m"))

	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	limiter := medjed.NewSlidingWindowLimiter(redisAddr, rateLimit, rateWindow)

	app := gin.Default()
	app.Use(middleware.RateLimitMiddleware(limiter))

	app.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	app.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: app.Handler(),
	}

	go func() {
		log.Printf("Starting server on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
