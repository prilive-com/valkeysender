package valkeysender

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

// valkeySender implements the Sender interface using Redis Lists
type valkeySender struct {
	config     *Config
	client     *redis.Client
	logger     *slog.Logger
	options    *SenderOptions
	serializer MessageSerializer
	
	// Circuit breaker and rate limiter
	circuitBreaker *gobreaker.CircuitBreaker
	rateLimiter    *rate.Limiter
	
	// Metrics and health
	startTime      time.Time
	messagesSent   int64
	errorCount     int64
	lastSuccess    time.Time
	lastError      string
	isConnected    bool
	connectionMutex sync.RWMutex
	
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSender creates a new Valkey sender
func NewSender(config *Config, options *SenderOptions) (Sender, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	if options == nil {
		options = &SenderOptions{}
	}
	
	// Create logger if not provided
	var logger *slog.Logger
	if options.Logger != nil {
		if l, ok := options.Logger.(*slog.Logger); ok {
			logger = l
		} else {
			return nil, fmt.Errorf("logger must be of type *slog.Logger")
		}
	}
	
	if logger == nil {
		var err error
		logger, err = NewLogger(config.LogSlogLevel(), "")
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
	}
	
	// Create serializer if not provided
	serializer := options.Serializer
	if serializer == nil {
		serializer = NewJSONSerializer()
	}
	
	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	sender := &valkeySender{
		config:     config,
		logger:     logger,
		options:    options,
		serializer: serializer,
		startTime:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// Initialize circuit breaker
	sender.circuitBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "valkeysender",
		MaxRequests: config.BreakerMaxRequests,
		Interval:    config.BreakerInterval,
		Timeout:     config.BreakerTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			sender.logger.Info("Circuit breaker state change",
				slog.String("name", name),
				slog.String("from", from.String()),
				slog.String("to", to.String()),
			)
		},
	})
	
	// Initialize rate limiter
	sender.rateLimiter = rate.NewLimiter(rate.Limit(config.RateLimitRequests), config.RateLimitBurst)
	
	// Initialize Redis client
	if err := sender.initClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}
	
	// Test connection
	if err := sender.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}
	
	sender.logger.Info("Valkey sender created",
		slog.String("address", config.Address),
		slog.Int("database", config.Database),
		slog.String("default_queue", config.DefaultQueue),
	)
	
	return sender, nil
}

// initClient initializes the Redis client with proper configuration
func (s *valkeySender) initClient() error {
	opts := &redis.Options{
		Addr:         s.config.Address,
		Username:     s.config.Username,
		Password:     s.config.Password,
		DB:           s.config.Database,
		DialTimeout:  s.config.DialTimeout,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		PoolSize:     s.config.PoolSize,
		MinIdleConns: s.config.MinIdleConns,
		ConnMaxIdleTime: s.config.MaxIdleTime,
		ConnMaxLifetime: s.config.ConnMaxLifetime,
	}
	
	// Configure TLS if enabled
	if s.config.TLSEnabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: s.config.TLSSkipVerify,
		}
		
		if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
			cert, err := tls.LoadX509KeyPair(s.config.TLSCertFile, s.config.TLSKeyFile)
			if err != nil {
				return fmt.Errorf("failed to load TLS certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		
		opts.TLSConfig = tlsConfig
	}
	
	s.client = redis.NewClient(opts)
	return nil
}

// testConnection tests the connection to Valkey
func (s *valkeySender) testConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.DialTimeout)
	defer cancel()
	
	// Test basic connectivity
	pong, err := s.client.Ping(ctx).Result()
	if err != nil {
		s.setConnectionState(false)
		return fmt.Errorf("failed to ping Valkey: %w", err)
	}
	
	if pong != "PONG" {
		s.setConnectionState(false)
		return fmt.Errorf("unexpected ping response: %s", pong)
	}
	
	s.setConnectionState(true)
	s.lastSuccess = time.Now()
	
	s.logger.Info("Successfully connected to Valkey",
		slog.String("address", s.config.Address),
		slog.Int("database", s.config.Database),
	)
	
	return nil
}

// setConnectionState updates the connection state thread-safely
func (s *valkeySender) setConnectionState(connected bool) {
	s.connectionMutex.Lock()
	defer s.connectionMutex.Unlock()
	s.isConnected = connected
}

// getConnectionState gets the connection state thread-safely
func (s *valkeySender) getConnectionState() bool {
	s.connectionMutex.RLock()
	defer s.connectionMutex.RUnlock()
	return s.isConnected
}

// SendMessage sends a message to the specified queue
func (s *valkeySender) SendMessage(ctx context.Context, queue string, message interface{}) error {
	return s.SendMessageWithTTL(ctx, queue, message, s.config.MessageTTL)
}

// SendMessageWithTTL sends a message with custom TTL
func (s *valkeySender) SendMessageWithTTL(ctx context.Context, queue string, message interface{}, ttl time.Duration) error {
	startTime := time.Now()
	
	// Apply rate limiting
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}
	
	// Use circuit breaker
	_, err := s.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, s.sendMessageInternal(ctx, queue, message, ttl)
	})
	
	if err != nil {
		atomic.AddInt64(&s.errorCount, 1)
		s.lastError = err.Error()
		
		if s.options.ErrorHandler != nil {
			s.options.ErrorHandler(err)
		}
		
		return err
	}
	
	// Update metrics
	atomic.AddInt64(&s.messagesSent, 1)
	s.lastSuccess = time.Now()
	
	// Call success handler
	if s.options.SuccessHandler != nil {
		metadata := MessageMetadata{
			Queue:     queue,
			MessageID: uuid.New().String(),
			Timestamp: startTime,
			TTL:       ttl,
		}
		s.options.SuccessHandler(metadata)
	}
	
	return nil
}

// sendMessageInternal performs the actual message sending
func (s *valkeySender) sendMessageInternal(ctx context.Context, queue string, message interface{}, ttl time.Duration) error {
	// Create message envelope
	envelope := MessageEnvelope{
		ID:        uuid.New().String(),
		Queue:     queue,
		Timestamp: time.Now(),
		TTL:       ttl,
		Headers:   make(map[string]string),
	}
	
	// Serialize the message payload
	payload, err := s.serializer.Serialize(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}
	envelope.Payload = payload
	
	// Serialize the envelope
	envelopeData, err := SerializeMessageEnvelope(envelope)
	if err != nil {
		return fmt.Errorf("failed to serialize envelope: %w", err)
	}
	
	// Send to Redis List using LPUSH (add to left side)
	listKey := s.getQueueKey(queue)
	
	pipe := s.client.Pipeline()
	
	// Add message to list
	pipe.LPush(ctx, listKey, envelopeData)
	
	// Set TTL on the list itself if it doesn't exist
	pipe.Expire(ctx, listKey, ttl)
	
	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		s.setConnectionState(false)
		return fmt.Errorf("failed to send message to queue %s: %w", queue, err)
	}
	
	s.setConnectionState(true)
	
	s.logger.Debug("Message sent successfully",
		slog.String("queue", queue),
		slog.String("message_id", envelope.ID),
		slog.Int("payload_size", len(payload)),
		slog.Duration("ttl", ttl),
	)
	
	return nil
}


// SendBatch sends multiple messages to the same queue atomically
func (s *valkeySender) SendBatch(ctx context.Context, queue string, messages []interface{}) error {
	if len(messages) == 0 {
		return fmt.Errorf("messages slice cannot be empty")
	}
	
	startTime := time.Now()
	
	// Apply rate limiting (once for the batch)
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}
	
	// Use circuit breaker
	_, err := s.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, s.sendBatchInternal(ctx, queue, messages)
	})
	
	if err != nil {
		atomic.AddInt64(&s.errorCount, 1)
		s.lastError = err.Error()
		
		if s.options.ErrorHandler != nil {
			s.options.ErrorHandler(err)
		}
		
		return err
	}
	
	// Update metrics
	atomic.AddInt64(&s.messagesSent, int64(len(messages)))
	s.lastSuccess = time.Now()
	
	// Call success handler for each message
	if s.options.SuccessHandler != nil {
		for i := range messages {
			metadata := MessageMetadata{
				Queue:     queue,
				Position:  int64(i),
				MessageID: uuid.New().String(),
				Timestamp: startTime,
				TTL:       s.config.MessageTTL,
			}
			s.options.SuccessHandler(metadata)
		}
	}
	
	return nil
}

// sendBatchInternal performs the actual batch message sending
func (s *valkeySender) sendBatchInternal(ctx context.Context, queue string, messages []interface{}) error {
	listKey := s.getQueueKey(queue)
	
	// Prepare all envelopes
	envelopes := make([]interface{}, len(messages))
	
	for i, message := range messages {
		envelope := MessageEnvelope{
			ID:        uuid.New().String(),
			Queue:     queue,
			Timestamp: time.Now(),
			TTL:       s.config.MessageTTL,
			Headers:   make(map[string]string),
		}
		
		// Serialize the message payload
		payload, err := s.serializer.Serialize(message)
		if err != nil {
			return fmt.Errorf("failed to serialize message %d: %w", i, err)
		}
		envelope.Payload = payload
		
		// Serialize the envelope
		envelopeData, err := SerializeMessageEnvelope(envelope)
		if err != nil {
			return fmt.Errorf("failed to serialize envelope %d: %w", i, err)
		}
		
		envelopes[i] = envelopeData
	}
	
	// Send all messages atomically using LPUSH
	pipe := s.client.Pipeline()
	
	// Add all messages to list
	pipe.LPush(ctx, listKey, envelopes...)
	
	// Set TTL on the list
	pipe.Expire(ctx, listKey, s.config.MessageTTL)
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		s.setConnectionState(false)
		return fmt.Errorf("failed to send batch to queue %s: %w", queue, err)
	}
	
	s.setConnectionState(true)
	
	s.logger.Debug("Batch sent successfully",
		slog.String("queue", queue),
		slog.Int("message_count", len(messages)),
	)
	
	return nil
}

// GetQueueSize returns the current size of a queue
func (s *valkeySender) GetQueueSize(ctx context.Context, queue string) (int64, error) {
	listKey := s.getQueueKey(queue)
	
	size, err := s.client.LLen(ctx, listKey).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Queue doesn't exist, so size is 0
		}
		return 0, fmt.Errorf("failed to get queue size for %s: %w", queue, err)
	}
	
	return size, nil
}

// Close gracefully shuts down the sender
func (s *valkeySender) Close() error {
	s.logger.Info("Closing Valkey sender")
	
	// Cancel context to stop all operations
	s.cancel()
	
	// Wait for all goroutines to finish
	s.wg.Wait()
	
	// Close Redis client
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			s.logger.Error("Error closing Redis client", slog.Any("error", err))
			return err
		}
	}
	
	s.setConnectionState(false)
	s.logger.Info("Valkey sender closed")
	
	return nil
}

// Health returns the health status of the sender
func (s *valkeySender) Health() HealthStatus {
	connectionState := "disconnected"
	if s.getConnectionState() {
		connectionState = "connected"
	}
	
	status := "healthy"
	errorRate := float64(atomic.LoadInt64(&s.errorCount)) / float64(atomic.LoadInt64(&s.messagesSent)+1)
	switch {
	case errorRate > 0.5:
		status = "unhealthy"
	case errorRate > 0.1:
		status = "degraded"
	}
	
	return HealthStatus{
		Status:          status,
		LastSuccess:     s.lastSuccess,
		LastError:       s.lastError,
		ErrorCount:      atomic.LoadInt64(&s.errorCount),
		MessagesSent:    atomic.LoadInt64(&s.messagesSent),
		Uptime:          time.Since(s.startTime),
		ConnectionState: connectionState,
		CircuitBreaker:  s.circuitBreaker.State().String(),
	}
}

// getQueueKey returns the Redis key for a queue
func (s *valkeySender) getQueueKey(queue string) string {
	if s.options.QueueNamer != nil {
		return s.options.QueueNamer(queue)
	}
	return fmt.Sprintf("queue:%s", queue)
}