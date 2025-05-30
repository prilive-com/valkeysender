package valkeysender

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	for _, env := range []string{
		"VALKEY_SENDER_ADDRESS",
		"VALKEY_SENDER_PASSWORD",
		"VALKEY_SENDER_DATABASE",
		"VALKEY_SENDER_DEFAULT_QUEUE",
		"VALKEY_SENDER_TLS_ENABLED",
		"VALKEY_SENDER_TLS_CERT_FILE",
		"VALKEY_SENDER_TLS_KEY_FILE",
	} {
		if val := os.Getenv(env); val != "" {
			originalEnv[env] = val
		}
		os.Unsetenv(env)
	}
	
	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			os.Setenv(key, value)
		}
	}()
	
	tests := []struct {
		name        string
		setupEnv    func()
		expectError bool
		validate    func(*Config) error
	}{
		{
			name: "default configuration",
			setupEnv: func() {
				// No environment variables set
			},
			expectError: false,
			validate: func(c *Config) error {
				if c.Address != "localhost:6379" {
					t.Errorf("Expected default address localhost:6379, got %s", c.Address)
				}
				if c.Database != 0 {
					t.Errorf("Expected default database 0, got %d", c.Database)
				}
				if c.DefaultQueue != "user-registrations" {
					t.Errorf("Expected default queue user-registrations, got %s", c.DefaultQueue)
				}
				return nil
			},
		},
		{
			name: "custom configuration",
			setupEnv: func() {
				os.Setenv("VALKEY_SENDER_ADDRESS", "redis.example.com:6380")
				os.Setenv("VALKEY_SENDER_PASSWORD", "secret123")
				os.Setenv("VALKEY_SENDER_DATABASE", "2")
				os.Setenv("VALKEY_SENDER_DEFAULT_QUEUE", "custom-queue")
			},
			expectError: false,
			validate: func(c *Config) error {
				if c.Address != "redis.example.com:6380" {
					t.Errorf("Expected address redis.example.com:6380, got %s", c.Address)
				}
				if c.Password != "secret123" {
					t.Errorf("Expected password secret123, got %s", c.Password)
				}
				if c.Database != 2 {
					t.Errorf("Expected database 2, got %d", c.Database)
				}
				if c.DefaultQueue != "custom-queue" {
					t.Errorf("Expected queue custom-queue, got %s", c.DefaultQueue)
				}
				return nil
			},
		},
		{
			name: "TLS configuration with missing files",
			setupEnv: func() {
				os.Setenv("VALKEY_SENDER_TLS_ENABLED", "true")
				// Missing cert and key files
			},
			expectError: true,
		},
		{
			name: "TLS configuration with files",
			setupEnv: func() {
				os.Setenv("VALKEY_SENDER_TLS_ENABLED", "true")
				os.Setenv("VALKEY_SENDER_TLS_CERT_FILE", "/path/to/cert.pem")
				os.Setenv("VALKEY_SENDER_TLS_KEY_FILE", "/path/to/key.pem")
			},
			expectError: false,
			validate: func(c *Config) error {
				if !c.TLSEnabled {
					t.Errorf("Expected TLS enabled")
				}
				if c.TLSCertFile != "/path/to/cert.pem" {
					t.Errorf("Expected cert file /path/to/cert.pem, got %s", c.TLSCertFile)
				}
				return nil
			},
		},
		{
			name: "invalid database number",
			setupEnv: func() {
				os.Setenv("VALKEY_SENDER_DATABASE", "16")
			},
			expectError: true,
		},
		{
			name: "timeout configurations",
			setupEnv: func() {
				os.Setenv("VALKEY_SENDER_DIAL_TIMEOUT", "10s")
				os.Setenv("VALKEY_SENDER_READ_TIMEOUT", "5s")
				os.Setenv("VALKEY_SENDER_WRITE_TIMEOUT", "5s")
				os.Setenv("VALKEY_SENDER_MESSAGE_TTL", "1h")
			},
			expectError: false,
			validate: func(c *Config) error {
				if c.DialTimeout != 10*time.Second {
					t.Errorf("Expected dial timeout 10s, got %v", c.DialTimeout)
				}
				if c.ReadTimeout != 5*time.Second {
					t.Errorf("Expected read timeout 5s, got %v", c.ReadTimeout)
				}
				if c.MessageTTL != time.Hour {
					t.Errorf("Expected message TTL 1h, got %v", c.MessageTTL)
				}
				return nil
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for _, env := range []string{
				"VALKEY_SENDER_ADDRESS",
				"VALKEY_SENDER_PASSWORD",
				"VALKEY_SENDER_DATABASE",
				"VALKEY_SENDER_DEFAULT_QUEUE",
				"VALKEY_SENDER_TLS_ENABLED",
				"VALKEY_SENDER_TLS_CERT_FILE",
				"VALKEY_SENDER_TLS_KEY_FILE",
				"VALKEY_SENDER_DIAL_TIMEOUT",
				"VALKEY_SENDER_READ_TIMEOUT",
				"VALKEY_SENDER_WRITE_TIMEOUT",
				"VALKEY_SENDER_MESSAGE_TTL",
			} {
				os.Unsetenv(env)
			}
			
			// Setup environment
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			
			// Load config
			config, err := LoadConfig()
			
			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}
			
			// Validate config
			if !tt.expectError && tt.validate != nil {
				if err := tt.validate(config); err != nil {
					t.Errorf("Config validation failed: %v", err)
				}
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Address:      "localhost:6379",
				Database:     0,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				MinIdleConns: 2,
				DefaultQueue: "test-queue",
				MessageTTL:   24 * time.Hour,
				MaxRetries:   3,
				RetryDelay:   time.Second,
			},
			expectError: false,
		},
		{
			name: "empty address",
			config: &Config{
				Address: "",
			},
			expectError: true,
		},
		{
			name: "invalid database",
			config: &Config{
				Address:  "localhost:6379",
				Database: -1,
			},
			expectError: true,
		},
		{
			name: "invalid database too high",
			config: &Config{
				Address:  "localhost:6379",
				Database: 16,
			},
			expectError: true,
		},
		{
			name: "zero dial timeout",
			config: &Config{
				Address:     "localhost:6379",
				Database:    0,
				DialTimeout: 0,
			},
			expectError: true,
		},
		{
			name: "invalid pool size",
			config: &Config{
				Address:      "localhost:6379",
				Database:     0,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     0,
			},
			expectError: true,
		},
		{
			name: "min idle conns exceeds pool size",
			config: &Config{
				Address:      "localhost:6379",
				Database:     0,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     5,
				MinIdleConns: 10,
			},
			expectError: true,
		},
		{
			name: "empty queue name",
			config: &Config{
				Address:      "localhost:6379",
				Database:     0,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				MinIdleConns: 2,
				DefaultQueue: "",
			},
			expectError: true,
		},
		{
			name: "short message TTL",
			config: &Config{
				Address:      "localhost:6379",
				Database:     0,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				MinIdleConns: 2,
				DefaultQueue: "test-queue",
				MessageTTL:   500 * time.Millisecond,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}