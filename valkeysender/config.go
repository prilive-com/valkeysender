package valkeysender

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Core Valkey/Redis settings
	Address  string
	Username string
	Password string
	Database int
	
	// Connection settings
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	PoolSize       int
	MinIdleConns   int
	MaxIdleTime    time.Duration
	ConnMaxLifetime time.Duration
	
	// Message settings
	DefaultQueue   string
	MessageTTL     time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	
	// Circuit breaker settings
	BreakerMaxRequests uint32
	BreakerInterval    time.Duration
	BreakerTimeout     time.Duration
	
	// Rate limiting
	RateLimitRequests int
	RateLimitBurst    int
	
	// TLS settings
	TLSEnabled     bool
	TLSSkipVerify  bool
	TLSCertFile    string
	TLSKeyFile     string
	TLSCAFile      string
	
	// Logging
	LogLevel string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		// Default values
		Address:         getEnvOrDefault("VALKEY_SENDER_ADDRESS", "localhost:6379"),
		Username:        os.Getenv("VALKEY_SENDER_USERNAME"),
		Password:        os.Getenv("VALKEY_SENDER_PASSWORD"),
		Database:        parseIntOrDefault("VALKEY_SENDER_DATABASE", "0"),
		DialTimeout:     parseDurationOrDefault("VALKEY_SENDER_DIAL_TIMEOUT", "5s"),
		ReadTimeout:     parseDurationOrDefault("VALKEY_SENDER_READ_TIMEOUT", "3s"),
		WriteTimeout:    parseDurationOrDefault("VALKEY_SENDER_WRITE_TIMEOUT", "3s"),
		PoolSize:        parseIntOrDefault("VALKEY_SENDER_POOL_SIZE", "10"),
		MinIdleConns:    parseIntOrDefault("VALKEY_SENDER_MIN_IDLE_CONNS", "2"),
		MaxIdleTime:     parseDurationOrDefault("VALKEY_SENDER_MAX_IDLE_TIME", "5m"),
		ConnMaxLifetime: parseDurationOrDefault("VALKEY_SENDER_CONN_MAX_LIFETIME", "1h"),
		DefaultQueue:    getEnvOrDefault("VALKEY_SENDER_DEFAULT_QUEUE", "user-registrations"),
		MessageTTL:      parseDurationOrDefault("VALKEY_SENDER_MESSAGE_TTL", "24h"),
		MaxRetries:      parseIntOrDefault("VALKEY_SENDER_MAX_RETRIES", "3"),
		RetryDelay:      parseDurationOrDefault("VALKEY_SENDER_RETRY_DELAY", "1s"),
		BreakerMaxRequests: parseUint32OrDefault("VALKEY_SENDER_BREAKER_MAX_REQUESTS", "5"),
		BreakerInterval:    parseDurationOrDefault("VALKEY_SENDER_BREAKER_INTERVAL", "2m"),
		BreakerTimeout:     parseDurationOrDefault("VALKEY_SENDER_BREAKER_TIMEOUT", "60s"),
		RateLimitRequests:  parseIntOrDefault("VALKEY_SENDER_RATE_LIMIT_REQUESTS", "1000"),
		RateLimitBurst:     parseIntOrDefault("VALKEY_SENDER_RATE_LIMIT_BURST", "2000"),
		TLSEnabled:         parseBoolOrDefault("VALKEY_SENDER_TLS_ENABLED", "false"),
		TLSSkipVerify:      parseBoolOrDefault("VALKEY_SENDER_TLS_SKIP_VERIFY", "false"),
		TLSCertFile:        os.Getenv("VALKEY_SENDER_TLS_CERT_FILE"),
		TLSKeyFile:         os.Getenv("VALKEY_SENDER_TLS_KEY_FILE"),
		TLSCAFile:          os.Getenv("VALKEY_SENDER_TLS_CA_FILE"),
		LogLevel:           getEnvOrDefault("VALKEY_SENDER_LOG_LEVEL", "INFO"),
	}
	
	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	return config, nil
}

func (c *Config) validate() error {
	if c.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	
	if c.Database < 0 || c.Database > 15 {
		return fmt.Errorf("database must be between 0 and 15")
	}
	
	if c.DialTimeout < time.Millisecond {
		return fmt.Errorf("dial timeout must be at least 1ms")
	}
	
	if c.ReadTimeout < time.Millisecond {
		return fmt.Errorf("read timeout must be at least 1ms")
	}
	
	if c.WriteTimeout < time.Millisecond {
		return fmt.Errorf("write timeout must be at least 1ms")
	}
	
	if c.PoolSize < 1 {
		return fmt.Errorf("pool size must be at least 1")
	}
	
	if c.MinIdleConns < 0 {
		return fmt.Errorf("min idle connections cannot be negative")
	}
	
	if c.MinIdleConns > c.PoolSize {
		return fmt.Errorf("min idle connections cannot exceed pool size")
	}
	
	if c.DefaultQueue == "" {
		return fmt.Errorf("default queue name cannot be empty")
	}
	
	if c.MessageTTL < time.Second {
		return fmt.Errorf("message TTL must be at least 1 second")
	}
	
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	
	if c.RetryDelay < time.Millisecond {
		return fmt.Errorf("retry delay must be at least 1ms")
	}
	
	// TLS validation
	if c.TLSEnabled {
		if c.TLSCertFile == "" || c.TLSKeyFile == "" {
			return fmt.Errorf("TLS cert file and key file are required when TLS is enabled")
		}
	}
	
	return nil
}

func (c *Config) LogSlogLevel() slog.Level {
	switch strings.ToUpper(c.LogLevel) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDurationOrDefault(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

func parseIntOrDefault(key, defaultValue string) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	intVal, _ := strconv.Atoi(defaultValue)
	return intVal
}

func parseUint32OrDefault(key, defaultValue string) uint32 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseUint(value, 10, 32); err == nil {
			return uint32(intVal)
		}
	}
	intVal, _ := strconv.ParseUint(defaultValue, 10, 32)
	return uint32(intVal)
}

func parseBoolOrDefault(key, defaultValue string) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	boolVal, _ := strconv.ParseBool(defaultValue)
	return boolVal
}