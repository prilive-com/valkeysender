# valkeysender

**valkeysender** is a production-ready, high-performance Go library for sending messages to Valkey/Redis using Redis Lists with built-in resilience, observability, and performance optimization features.

## ✨ Features

| Capability | Details |
|------------|---------|
| **Production Ready** | Battle-tested patterns with circuit breakers, rate limiting, and retry logic |
| **High Performance** | Connection pooling, efficient batching, Redis Lists for reliable queuing |
| **Resilient** | Circuit breaker pattern, exponential backoff, graceful degradation |
| **Observable** | Structured logging (slog), health checks, performance metrics |
| **Secure** | TLS support, authentication, configurable timeouts |
| **Configurable** | Environment-based configuration with validation and sensible defaults |
| **Type Safe** | Clean interfaces, comprehensive error handling, context support |

## 📦 Installation

```bash
go get github.com/prilive-com/valkeysender
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/prilive-com/valkeysender/valkeysender"
)

func main() {
    // Load configuration from environment variables
    config, err := valkeysender.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Create sender
    sender, err := valkeysender.NewSender(config, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer sender.Close()

    // Send a message
    ctx := context.Background()
    err = sender.SendMessage(ctx, "my-queue", "Hello, Valkey!")
    if err != nil {
        log.Printf("Failed to send message: %v", err)
    }
}
```

### Environment Configuration

Set these environment variables:

```bash
# Required
VALKEY_SENDER_ADDRESS=localhost:6379

# Optional (with defaults)
VALKEY_SENDER_DATABASE=0
VALKEY_SENDER_DEFAULT_QUEUE=user-registrations
VALKEY_SENDER_MESSAGE_TTL=24h
VALKEY_SENDER_LOG_LEVEL=INFO
```

For authentication:
```bash
VALKEY_SENDER_USERNAME=your-username
VALKEY_SENDER_PASSWORD=your-password
```

For TLS:
```bash
VALKEY_SENDER_TLS_ENABLED=true
VALKEY_SENDER_TLS_CERT_FILE=/path/to/cert.pem
VALKEY_SENDER_TLS_KEY_FILE=/path/to/key.pem
```

## 🏗️ Architecture

```
┌──────────────┐   Redis Lists   ┌──────────────┐   BRPOP     ┌──────────────┐
│Your App Logic│ ─────────────▶  │    Valkey    │ ──────────▶ │   Consumer   │
└──────────────┘   LPUSH         │  (queue:*)   │  blocking   └──────────────┘
                                 │ (rate‑limit) │
                                 │ (circuit‑br) │
                                 │ (retry)      │
                                 └──────────────┘
```

### Why Redis Lists?

- ✅ **Reliable**: LPUSH/BRPOP operations are atomic and durable
- ✅ **Simple**: Easy to understand and debug with Redis CLI
- ✅ **Persistent**: Messages survive Valkey restarts (with proper persistence)
- ✅ **Blocking**: BRPOP waits efficiently for new messages
- ✅ **FIFO**: Messages processed in first-in-first-out order
- ✅ **Multiple consumers**: Multiple services can consume from the same queue

### Core Components

- **Config**: Environment-based configuration with validation
- **Sender**: Main message sending interface with resilience features
- **Serializer**: JSON message serialization with custom options
- **Logger**: Structured logging with performance insights
- **Circuit Breaker**: Prevents cascade failures using sony/gobreaker
- **Rate Limiter**: Token bucket rate limiting using golang.org/x/time

## 📋 Configuration Reference

### Core Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `VALKEY_SENDER_ADDRESS` | `localhost:6379` | Valkey/Redis server address |
| `VALKEY_SENDER_DATABASE` | `0` | Database number (0-15) |
| `VALKEY_SENDER_DEFAULT_QUEUE` | `user-registrations` | Default queue name |

### Connection Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `VALKEY_SENDER_DIAL_TIMEOUT` | `5s` | Connection timeout |
| `VALKEY_SENDER_READ_TIMEOUT` | `3s` | Read operation timeout |
| `VALKEY_SENDER_WRITE_TIMEOUT` | `3s` | Write operation timeout |
| `VALKEY_SENDER_POOL_SIZE` | `10` | Maximum connections in pool |
| `VALKEY_SENDER_MIN_IDLE_CONNS` | `2` | Minimum idle connections |
| `VALKEY_SENDER_MAX_IDLE_TIME` | `5m` | Maximum idle time for connections |
| `VALKEY_SENDER_CONN_MAX_LIFETIME` | `1h` | Maximum lifetime for connections |

### Message Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `VALKEY_SENDER_MESSAGE_TTL` | `24h` | Default message time-to-live |
| `VALKEY_SENDER_MAX_RETRIES` | `3` | Maximum retry attempts |
| `VALKEY_SENDER_RETRY_DELAY` | `1s` | Delay between retries |

### Security

| Variable | Default | Description |
|----------|---------|-------------|
| `VALKEY_SENDER_USERNAME` | | Username for authentication |
| `VALKEY_SENDER_PASSWORD` | | Password for authentication |
| `VALKEY_SENDER_TLS_ENABLED` | `false` | Enable TLS/SSL |
| `VALKEY_SENDER_TLS_CERT_FILE` | | TLS certificate file |
| `VALKEY_SENDER_TLS_KEY_FILE` | | TLS private key file |
| `VALKEY_SENDER_TLS_CA_FILE` | | TLS CA certificate file |
| `VALKEY_SENDER_TLS_SKIP_VERIFY` | `false` | Skip certificate verification |

### Resilience

| Variable | Default | Description |
|----------|---------|-------------|
| `VALKEY_SENDER_RATE_LIMIT_REQUESTS` | `1000` | Requests per second limit |
| `VALKEY_SENDER_RATE_LIMIT_BURST` | `2000` | Burst token bucket size |
| `VALKEY_SENDER_BREAKER_MAX_REQUESTS` | `5` | Circuit breaker half-open requests |
| `VALKEY_SENDER_BREAKER_INTERVAL` | `2m` | Circuit breaker reset interval |
| `VALKEY_SENDER_BREAKER_TIMEOUT` | `60s` | Circuit breaker open timeout |

### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `VALKEY_SENDER_LOG_LEVEL` | `INFO` | Log level (DEBUG, INFO, WARN, ERROR) |

## 🔧 Advanced Usage

### Custom Options

```go
options := &valkeysender.SenderOptions{
    Logger: logger,
    
    ErrorHandler: func(err error) {
        log.Printf("Sender error: %v", err)
    },
    
    SuccessHandler: func(metadata valkeysender.MessageMetadata) {
        log.Printf("Message sent: queue=%s id=%s", metadata.Queue, metadata.MessageID)
    },
    
    // Custom queue naming strategy
    QueueNamer: func(queue string) string {
        return fmt.Sprintf("myapp:queue:%s", queue)
    },
    
    // Custom serializer
    Serializer: customSerializer,
}

sender, err := valkeysender.NewSender(config, options)
```

### Sending User Registration Data

```go
userData := valkeysender.UserRegistrationData{
    Name:             "John Doe",
    Email:            "john@example.com",
    TelegramUserID:   123456789,
    TelegramUsername: "johndoe",
    FirstName:        "John",
    LastName:         "Doe",
    PhoneNumber:      "+1234567890",
    LanguageCode:     "en",
    Source:           "telegram-bot",
}

err := sender.SendUserRegistration(ctx, "user-registrations", userData)
```

### Sending with Custom TTL

```go
// Message expires in 30 minutes
err := sender.SendMessageWithTTL(ctx, "temp-queue", "urgent message", 30*time.Minute)
```

### Batch Operations

```go
messages := []interface{}{
    "Message 1",
    "Message 2",
    userData,
}

err := sender.SendBatch(ctx, "batch-queue", messages)
```

### Queue Monitoring

```go
size, err := sender.GetQueueSize(ctx, "user-registrations")
if err != nil {
    log.Printf("Queue has %d messages", size)
}
```

### Health Monitoring

```go
health := sender.Health()
fmt.Printf("Status: %s\n", health.Status)         // healthy, degraded, unhealthy
fmt.Printf("Messages Sent: %d\n", health.MessagesSent)
fmt.Printf("Error Count: %d\n", health.ErrorCount)
fmt.Printf("Uptime: %v\n", health.Uptime)
fmt.Printf("Connection: %s\n", health.ConnectionState)
fmt.Printf("Circuit Breaker: %s\n", health.CircuitBreaker)
```

## 🧪 Testing

Run the test suite:

```bash
cd valkeysender
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

Run the simple test (requires local Redis/Valkey):

```bash
go run test_simple.go
```

Run the full demo:

```bash
go run ./example/
```

## 📊 Performance

### Typical Performance

- **Throughput**: 10,000+ messages/second (depending on message size and Valkey setup)
- **Latency**: Sub-millisecond for LPUSH operations
- **Memory**: Efficient connection pooling and message batching
- **CPU**: Optimized Redis protocol usage

### Optimization Tips

1. **Use batch operations** for multiple messages
2. **Tune connection pool** settings based on load
3. **Monitor circuit breaker** and rate limiter metrics
4. **Use appropriate TTL** values for your use case
5. **Enable connection pooling** for high-throughput scenarios

## 🐳 Docker Support

Example Dockerfile:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o valkeysender-app ./example

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/valkeysender-app .
CMD ["./valkeysender-app"]
```

Example docker-compose.yml:

```yaml
version: '3.8'
services:
  valkeysender:
    build: .
    environment:
      - VALKEY_SENDER_ADDRESS=valkey:6379
      - VALKEY_SENDER_DEFAULT_QUEUE=user-registrations
      - VALKEY_SENDER_LOG_LEVEL=INFO
    depends_on:
      - valkey
      
  valkey:
    image: valkey/valkey:8-alpine
    ports:
      - "6379:6379"
```

## 🔍 Monitoring

### Structured Logs

All logs are JSON formatted with consistent fields:

```json
{
  "timestamp": "2025-05-28T10:30:00Z",
  "level": "INFO",
  "component": "valkeysender",
  "msg": "Message sent successfully",
  "queue": "user-registrations",
  "message_id": "uuid-123",
  "ttl": "24h"
}
```

### Redis CLI Monitoring

Monitor your queues using Redis CLI:

```bash
# Check queue size
redis-cli LLEN queue:user-registrations

# Inspect messages (without consuming)
redis-cli LRANGE queue:user-registrations 0 9

# Consume a message (blocking)
redis-cli BRPOP queue:user-registrations 0

# Monitor queue activity in real-time
redis-cli --latency-history -i 1

# Get memory usage
redis-cli MEMORY USAGE queue:user-registrations
```

## 🎯 Use Cases

### Telegram Bot User Registration

Perfect for queuing user registrations from your Telegram dispatcher:

```go
// In your dispatcher
sender.SendUserRegistration(ctx, "user-registrations", userData)

// Consumer service (using Redis CLI or custom consumer)
redis-cli BRPOP queue:user-registrations 0
```

### Task Queue System

Use as a reliable task queue:

```go
type Task struct {
    ID      string    `json:"id"`
    Type    string    `json:"type"`
    Payload []byte    `json:"payload"`
    Created time.Time `json:"created"`
}

task := Task{
    ID:      uuid.New().String(),
    Type:    "send-email",
    Payload: emailData,
    Created: time.Now(),
}

sender.SendMessage(ctx, "email-tasks", task)
```

### Event Streaming

Stream events between microservices:

```go
event := map[string]interface{}{
    "event_type": "user_registered",
    "user_id":    123,
    "timestamp":  time.Now(),
    "metadata":   metadata,
}

sender.SendMessage(ctx, "events", event)
```

## 🔄 Migration from Kafka

### Advantages over Kafka

| **Aspect** | **Kafka** | **Valkey** |
|------------|-----------|------------|
| **Setup Complexity** | Complex (brokers, topics, ACLs) | Simple (single instance) |
| **Authentication** | SASL, complex permissions | Simple auth or no auth |
| **Operations** | Heavy cluster management | Lightweight, easy ops |
| **Debugging** | Complex tooling required | Redis CLI built-in |
| **Latency** | Higher (network overhead) | Lower (direct connection) |
| **Resource Usage** | Heavy (JVM, multiple processes) | Light (single process) |

### Migration Strategy

1. **Replace kafkasender** with valkeysender in your application
2. **Update configuration** from Kafka env vars to Valkey env vars
3. **Change consumer** from Kafka consumer to Redis BRPOP
4. **Test thoroughly** with your message volume

### Code Changes Required

```go
// Before (Kafka)
kafkaProducer.SendMessage(ctx, "topic", "key", message)

// After (Valkey)
valkeySender.SendMessage(ctx, "queue", message)
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for your changes
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

MIT © 2025 Prilive Com

## 🙏 Acknowledgments

- Built on top of [go-redis/redis](https://github.com/redis/go-redis)
- Circuit breaker by [Sony GoBreaker](https://github.com/sony/gobreaker)
- Rate limiting by [golang.org/x/time](https://pkg.go.dev/golang.org/x/time/rate)
- Inspired by the simplicity and reliability of Redis Lists

## 🔗 Related Projects

- [valkeyreceiver](../valkeyreceiver) - Valkey consumer library with same architecture
- [kafkasender](../kafkasender) - Kafka producer library (being replaced)
- [telegramreceiver](../telegramreceiver) - Telegram webhook receiver
- [telegramsender](../telegramsender) - Telegram message sender