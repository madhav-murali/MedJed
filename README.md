# MedJed

A distributed rate limiter built in Go using Redis with a sliding window algorithm.

![CI](https://github.com/madhav-murali/MedJed/actions/workflows/ci.yml/badge.svg)

## Features

- **Sliding window** rate limiting via atomic Redis Lua scripts
- **Distributed** — works across multiple instances via shared Redis
- **Plug-and-play Gin middleware** — drop into any Gin project
- **Configurable** via environment variables

## Quickstart

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`.

## Use as a Library

Add MedJed to any Gin project:

```bash
go get github.com/madhav-murali/medjed
```

```go
import (
    "time"
    "github.com/madhav-murali/medjed"
    "github.com/madhav-murali/medjed/middleware"
)

func main() {
    limiter := medjed.NewSlidingWindowLimiter("localhost:6379", 100, time.Minute)

    router := gin.Default()
    router.Use(middleware.RateLimitMiddleware(limiter))
    // ... your routes
    router.Run(":8080")
}
```

You can also implement the `medjed.Limiter` interface to use a custom backend.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `REDIS_ADDR` | `localhost:6379` | Redis server address |
| `PORT` | `8080` | HTTP listen port |
| `RATE_LIMIT` | `100` | Max requests per window |
| `RATE_WINDOW` | `1m` | Window duration (Go duration format) |
| `GIN_MODE` | `debug` | Gin mode (`debug`, `release`) |

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Demo endpoint |
| `GET` | `/health` | Health check |

All endpoints are rate-limited. Exceeding the limit returns `429 Too Many Requests`.

## Development

```bash
# Run tests
go test -race -v ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Build
go build -o medjed ./cmd/main.go
```

## Architecture

```
Client → Gin + RateLimit Middleware → Handler
                 ↓
         Redis (Sorted Set + Lua)
```

Each request is scored by timestamp in a Redis sorted set. The Lua script atomically prunes expired entries, checks the count against the limit, and inserts the new request — all in a single round trip.
