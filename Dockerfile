FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o medjed ./cmd/main.go

FROM alpine:3.21

RUN adduser -D -u 1000 appuser
USER appuser

COPY --from=builder /app/medjed /usr/local/bin/medjed

EXPOSE 8080

ENTRYPOINT ["medjed"]
