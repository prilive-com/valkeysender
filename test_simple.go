package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prilive-com/valkeysender/valkeysender"
)

func main() {
	fmt.Println("🧪 Simple Valkey Sender Test")
	fmt.Println("============================")

	// Create minimal config for testing
	config := &valkeysender.Config{
		Address:           "10.1.0.4:30379",
		Password:          "7Xwdz01BYEu6p74sNRHf8He2",
		Database:          0,
		DefaultQueue:      "test-queue",
		MessageTTL:        24 * time.Hour,
		DialTimeout:       5 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      3 * time.Second,
		PoolSize:          10,
		MinIdleConns:      2,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		RateLimitRequests: 100,
		RateLimitBurst:    200,
		LogLevel:          "INFO",
	}

	fmt.Printf("Testing connection to %s (database %d)\n", config.Address, config.Database)

	sender, err := valkeysender.NewSender(config, nil)
	if err != nil {
		log.Fatalf("❌ Failed to create sender: %v", err)
	}
	defer sender.Close()

	fmt.Println("✅ Connected to Valkey successfully!")

	ctx := context.Background()
	testMessage := "Hello from valkeysender test!"

	fmt.Printf("📤 Sending test message: %s\n", testMessage)
	err = sender.SendMessage(ctx, "test-queue", testMessage)
	if err != nil {
		log.Fatalf("❌ Failed to send message: %v", err)
	}

	fmt.Println("✅ Message sent successfully!")

	// Check queue size
	size, err := sender.GetQueueSize(ctx, "test-queue")
	if err != nil {
		log.Printf("⚠️ Could not get queue size: %v", err)
	} else {
		fmt.Printf("📊 Queue size: %d messages\n", size)
	}

	// Show health
	health := sender.Health()
	fmt.Printf("💚 Health: %s (sent: %d, errors: %d)\n", 
		health.Status, health.MessagesSent, health.ErrorCount)

	fmt.Println("🎉 Test completed successfully!")
}