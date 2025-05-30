package valkeysender

import (
	"context"
	"time"
)

// Sender defines the interface for sending messages to Valkey
type Sender interface {
	// SendMessage sends a message to the specified queue
	SendMessage(ctx context.Context, queue string, message interface{}) error
	
	// SendMessageWithTTL sends a message with custom TTL
	SendMessageWithTTL(ctx context.Context, queue string, message interface{}, ttl time.Duration) error
	
	// SendUserRegistration is a convenience method for sending user registration data
	SendUserRegistration(ctx context.Context, queue string, userData UserRegistrationData) error
	
	// SendBatch sends multiple messages to the same queue atomically
	SendBatch(ctx context.Context, queue string, messages []interface{}) error
	
	// GetQueueSize returns the current size of a queue
	GetQueueSize(ctx context.Context, queue string) (int64, error)
	
	// Close gracefully shuts down the sender
	Close() error
	
	// Health returns the health status of the sender
	Health() HealthStatus
}

// UserRegistrationData represents user registration information
type UserRegistrationData struct {
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	TelegramUserID   int64     `json:"telegram_user_id"`
	TelegramUsername string    `json:"telegram_username,omitempty"`
	FirstName        string    `json:"first_name,omitempty"`
	LastName         string    `json:"last_name,omitempty"`
	PhoneNumber      string    `json:"phone_number,omitempty"`
	LanguageCode     string    `json:"language_code,omitempty"`
	RegistrationTime time.Time `json:"registration_time"`
	Source           string    `json:"source"`
}

// MessageMetadata contains metadata about sent messages
type MessageMetadata struct {
	Queue      string            `json:"queue"`
	Position   int64             `json:"position"`        // Position in the list
	MessageID  string            `json:"message_id"`      // UUID for the message
	Headers    map[string]string `json:"headers,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	TTL        time.Duration     `json:"ttl"`
	Size       int               `json:"size"`            // Message size in bytes
}

// HealthStatus represents the health of the sender
type HealthStatus struct {
	Status          string        `json:"status"` // healthy, degraded, unhealthy
	LastSuccess     time.Time     `json:"last_success"`
	LastError       string        `json:"last_error,omitempty"`
	ErrorCount      int64         `json:"error_count"`
	MessagesSent    int64         `json:"messages_sent"`
	Uptime          time.Duration `json:"uptime"`
	ConnectionState string        `json:"connection_state"` // connected, disconnected, connecting
	CircuitBreaker  string        `json:"circuit_breaker"`  // closed, half-open, open
}

// SenderMetrics contains performance metrics
type SenderMetrics struct {
	MessagesSent        int64         `json:"messages_sent"`
	MessagesFailedTotal int64         `json:"messages_failed_total"`
	MessagesFailedLast  int64         `json:"messages_failed_last_minute"`
	AvgLatency          time.Duration `json:"avg_latency"`
	MaxLatency          time.Duration `json:"max_latency"`
	CircuitBreakerState string        `json:"circuit_breaker_state"`
	RateLimitHits       int64         `json:"rate_limit_hits"`
	QueueSizes          map[string]int64 `json:"queue_sizes"`
	ConnectionPool      PoolMetrics   `json:"connection_pool"`
	StartTime           time.Time     `json:"start_time"`
}

// PoolMetrics contains connection pool metrics
type PoolMetrics struct {
	TotalConns int32 `json:"total_conns"`
	IdleConns  int32 `json:"idle_conns"`
	StaleConns int32 `json:"stale_conns"`
	Hits       uint32 `json:"hits"`
	Misses     uint32 `json:"misses"`
	Timeouts   uint32 `json:"timeouts"`
}

// MessageResult represents the result of sending a message
type MessageResult struct {
	Success  bool              `json:"success"`
	Metadata *MessageMetadata  `json:"metadata,omitempty"`
	Error    error             `json:"error,omitempty"`
	Duration time.Duration     `json:"duration"`
}

// SenderOptions contains optional settings for creating a sender
type SenderOptions struct {
	// Custom error handler (optional)
	ErrorHandler func(error)
	
	// Custom success handler (optional)
	SuccessHandler func(MessageMetadata)
	
	// Custom metrics handler (optional)
	MetricsHandler func(SenderMetrics)
	
	// Logger for structured logging (if nil, a default logger will be created)
	Logger interface{}
	
	// Custom serializer (if nil, JSON will be used)
	Serializer MessageSerializer
	
	// Custom queue naming strategy
	QueueNamer func(queue string) string
	
	// Enable message deduplication
	EnableDeduplication bool
	
	// Deduplication window
	DeduplicationWindow time.Duration
}

// MessageSerializer defines the interface for message serialization
type MessageSerializer interface {
	Serialize(message interface{}) ([]byte, error)
	Deserialize(data []byte, target interface{}) error
	ContentType() string
}

// MessageEnvelope wraps messages with metadata for the queue
type MessageEnvelope struct {
	ID        string                 `json:"id"`
	Queue     string                 `json:"queue"`
	Payload   []byte                 `json:"payload"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	TTL       time.Duration          `json:"ttl"`
	Retries   int                    `json:"retries"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// QueueStats provides statistics about a queue
type QueueStats struct {
	Name           string        `json:"name"`
	Length         int64         `json:"length"`
	MemoryUsage    int64         `json:"memory_usage_bytes"`
	LastActivity   time.Time     `json:"last_activity"`
	MessagesPerSec float64       `json:"messages_per_sec"`
	AvgMessageSize float64       `json:"avg_message_size"`
}

// ConnectionInfo contains information about the Valkey connection
type ConnectionInfo struct {
	Address      string            `json:"address"`
	Database     int               `json:"database"`
	Username     string            `json:"username,omitempty"`
	TLSEnabled   bool              `json:"tls_enabled"`
	ServerInfo   map[string]string `json:"server_info,omitempty"`
	ClientInfo   map[string]string `json:"client_info,omitempty"`
	ConnectedAt  time.Time         `json:"connected_at"`
}

// BatchMessage represents a single message in a batch operation
type BatchMessage struct {
	Payload   interface{}       `json:"payload"`
	Headers   map[string]string `json:"headers,omitempty"`
	TTL       time.Duration     `json:"ttl,omitempty"`
}

// BatchResult represents the result of a batch send operation
type BatchResult struct {
	Success     bool                `json:"success"`
	TotalSent   int                 `json:"total_sent"`
	Failed      int                 `json:"failed"`
	Results     []MessageResult     `json:"results"`
	Duration    time.Duration       `json:"duration"`
	Error       error               `json:"error,omitempty"`
}