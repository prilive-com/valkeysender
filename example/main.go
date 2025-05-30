package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prilive-com/valkeysender/valkeysender"
)

func main() {
	fmt.Println("🚀 VALKEY SENDER DEMO")
	fmt.Println("=====================")
	fmt.Println()

	// Load configuration
	config, err := valkeysender.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Create logger
	logger, err := valkeysender.NewLogger(config.LogSlogLevel(), "logs/valkeysender-example.log")
	if err != nil {
		log.Fatalf("❌ Failed to create logger: %v", err)
	}

	logger.Info("Starting Valkey sender example",
		slog.Any("config", map[string]interface{}{
			"address":       config.Address,
			"database":      config.Database,
			"default_queue": config.DefaultQueue,
			"message_ttl":   config.MessageTTL,
		}),
	)

	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("   Address: %s\n", config.Address)
	fmt.Printf("   Database: %d\n", config.Database)
	fmt.Printf("   Default Queue: %s\n", config.DefaultQueue)
	fmt.Printf("   Message TTL: %s\n", config.MessageTTL)
	fmt.Printf("   TLS Enabled: %t\n", config.TLSEnabled)
	fmt.Println()

	// Create sender options with handlers
	options := &valkeysender.SenderOptions{
		Logger: logger,
		
		ErrorHandler: func(err error) {
			logger.Error("Sender error occurred", slog.Any("error", err))
			fmt.Printf("❌ ERROR: %v\n", err)
		},
		
		SuccessHandler: func(metadata valkeysender.MessageMetadata) {
			logger.Info("Message sent successfully",
				slog.String("queue", metadata.Queue),
				slog.String("message_id", metadata.MessageID),
				slog.Duration("ttl", metadata.TTL),
			)
			fmt.Printf("✅ SUCCESS: Message sent to queue=%s, id=%s\n", 
				metadata.Queue, metadata.MessageID)
		},
	}

	// Create sender
	sender, err := valkeysender.NewSender(config, options)
	if err != nil {
		log.Fatalf("❌ Failed to create sender: %v", err)
	}
	defer sender.Close()

	fmt.Println("✅ Valkey sender created successfully!")
	fmt.Println()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\n🛑 Shutdown signal received")
		cancel()
	}()

	// Demo different message types
	if err := runDemo(ctx, sender); err != nil {
		log.Fatalf("❌ Demo failed: %v", err)
	}

	// Show final health status
	health := sender.Health()
	fmt.Println()
	fmt.Println("📊 FINAL SENDER STATS")
	fmt.Println("═════════════════════")
	fmt.Printf("Status: %s\n", health.Status)
	fmt.Printf("Messages Sent: %d\n", health.MessagesSent)
	fmt.Printf("Errors: %d\n", health.ErrorCount)
	fmt.Printf("Uptime: %v\n", health.Uptime)
	fmt.Printf("Connection: %s\n", health.ConnectionState)
	fmt.Printf("Circuit Breaker: %s\n", health.CircuitBreaker)
	
	fmt.Println()
	fmt.Println("🎉 DEMO COMPLETED!")
	fmt.Println("═════════════════")
	fmt.Println("✅ All messages have been sent to Valkey")
	fmt.Println("🔍 Check your Valkey instance to see the messages")
	fmt.Println("📡 Messages are stored in Redis Lists and can be consumed with BRPOP")
}

func runDemo(ctx context.Context, sender valkeysender.Sender) error {
	// Test 1: Send a simple text message
	fmt.Println("📤 TEST 1: Sending simple text message")
	fmt.Println("─────────────────────────────────────")
	
	testMessage := "Hello from valkeysender demo! " + time.Now().Format("15:04:05")
	fmt.Printf("📝 Message content: %s\n", testMessage)
	fmt.Printf("🎯 Target queue: user-registrations\n")
	fmt.Println("⏳ Sending...")

	err := sender.SendMessage(ctx, "user-registrations", testMessage)
	if err != nil {
		return fmt.Errorf("failed to send simple message: %w", err)
	}
	fmt.Printf("✅ Simple message sent successfully!\n")
	fmt.Println()

	// Test 2: Send user registration data
	fmt.Println("📤 TEST 2: Sending user registration data")
	fmt.Println("─────────────────────────────────────")
	
	userData := valkeysender.UserRegistrationData{
		Name:             "John Demo User",
		Email:            "john.demo@example.com",
		TelegramUserID:   123456789,
		TelegramUsername: "johndemo",
		FirstName:        "John",
		LastName:         "Demo",
		PhoneNumber:      "+1234567890",
		LanguageCode:     "en",
		RegistrationTime: time.Now(),
		Source:           "valkeysender-demo",
	}

	// Pretty print the user data
	userDataJSON, _ := json.MarshalIndent(userData, "   ", "  ")
	fmt.Printf("📝 User registration data:\n   %s\n", string(userDataJSON))
	fmt.Printf("🎯 Target queue: user-registrations\n")
	fmt.Println("⏳ Sending...")

	err = sender.SendUserRegistration(ctx, "user-registrations", userData)
	if err != nil {
		return fmt.Errorf("failed to send user registration: %w", err)
	}
	fmt.Printf("✅ User registration data sent successfully!\n")
	fmt.Println()

	// Test 3: Send message with custom TTL
	fmt.Println("📤 TEST 3: Sending message with custom TTL")
	fmt.Println("──────────────────────────────────────────")
	
	shortLivedMessage := "This message expires in 30 seconds"
	fmt.Printf("📝 Message content: %s\n", shortLivedMessage)
	fmt.Printf("⏰ TTL: 30 seconds\n")
	fmt.Printf("🎯 Target queue: temporary-messages\n")
	fmt.Println("⏳ Sending...")

	err = sender.SendMessageWithTTL(ctx, "temporary-messages", shortLivedMessage, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to send TTL message: %w", err)
	}
	fmt.Printf("✅ TTL message sent successfully!\n")
	fmt.Println()

	// Test 4: Send batch messages
	fmt.Println("📤 TEST 4: Sending batch messages")
	fmt.Println("─────────────────────────────────")
	
	batchMessages := []interface{}{
		"Batch message #1 sent at " + time.Now().Format("15:04:05.000"),
		"Batch message #2 sent at " + time.Now().Format("15:04:05.000"),
		"Batch message #3 sent at " + time.Now().Format("15:04:05.000"),
	}

	fmt.Printf("📝 Sending %d messages in a batch\n", len(batchMessages))
	fmt.Printf("🎯 Target queue: batch-messages\n")
	fmt.Println("⏳ Sending...")

	err = sender.SendBatch(ctx, "batch-messages", batchMessages)
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}
	fmt.Printf("✅ Batch messages sent successfully!\n")
	fmt.Println()

	// Test 5: Check queue sizes
	fmt.Println("📊 TEST 5: Checking queue sizes")
	fmt.Println("──────────────────────────────")

	queues := []string{"user-registrations", "temporary-messages", "batch-messages"}
	for _, queue := range queues {
		size, err := sender.GetQueueSize(ctx, queue)
		if err != nil {
			fmt.Printf("❌ Failed to get size for queue %s: %v\n", queue, err)
			continue
		}
		fmt.Printf("📋 Queue '%s': %d messages\n", queue, size)
	}
	fmt.Println()

	fmt.Println("🎉 ALL TESTS COMPLETED!")
	fmt.Println("══════════════════════")
	fmt.Println("✅ Messages have been sent to Valkey using Redis Lists")
	fmt.Println("🔍 You can inspect the queues using Redis CLI:")
	fmt.Println("   redis-cli LLEN queue:user-registrations")
	fmt.Println("   redis-cli LRANGE queue:user-registrations 0 -1")
	fmt.Println("📡 Messages can be consumed using:")
	fmt.Println("   redis-cli BRPOP queue:user-registrations 0")

	return nil
}