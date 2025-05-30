package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/prilive-com/valkeysender/valkeysender"
)

func main() {
	fmt.Println("🧪 Valkey Sender Test (Environment Configuration)")
	fmt.Println("=================================================")

	// Load configuration from environment
	config, err := valkeysender.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Printf("📋 Configuration loaded:\n")
	fmt.Printf("   Address: %s\n", config.Address)
	fmt.Printf("   Database: %d\n", config.Database)
	fmt.Printf("   Default Queue: %s\n", config.DefaultQueue)
	fmt.Printf("   Message TTL: %s\n", config.MessageTTL)
	fmt.Println()

	fmt.Printf("🔗 Testing connection to %s (database %d)\n", config.Address, config.Database)

	sender, err := valkeysender.NewSender(config, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create sender: %v\n", err)
		fmt.Println()
		fmt.Println("💡 Possible solutions:")
		fmt.Println("   1. Start Redis/Valkey server:")
		fmt.Println("      docker run -d -p 6379:6379 redis:alpine")
		fmt.Println("   2. Or change VALKEY_SENDER_ADDRESS in .env file")
		fmt.Println("   3. Or set environment variable:")
		fmt.Println("      export VALKEY_SENDER_ADDRESS=your-server:6379")
		os.Exit(1)
	}
	defer sender.Close()

	fmt.Println("✅ Connected to Valkey successfully!")

	ctx := context.Background()
	testMessage := "Hello from valkeysender environment test!"

	fmt.Printf("📤 Sending test message: %s\n", testMessage)
	err = sender.SendMessage(ctx, config.DefaultQueue, testMessage)
	if err != nil {
		log.Fatalf("❌ Failed to send message: %v", err)
	}

	fmt.Println("✅ Message sent successfully!")

	// Check queue size
	size, err := sender.GetQueueSize(ctx, config.DefaultQueue)
	if err != nil {
		log.Printf("⚠️ Could not get queue size: %v", err)
	} else {
		fmt.Printf("📊 Queue '%s' size: %d messages\n", config.DefaultQueue, size)
	}

	// Show health
	health := sender.Health()
	fmt.Printf("💚 Health: %s (sent: %d, errors: %d, connection: %s)\n", 
		health.Status, health.MessagesSent, health.ErrorCount, health.ConnectionState)

	fmt.Println()
	fmt.Println("🎉 Test completed successfully!")
	fmt.Println("💡 You can inspect the queue using Redis CLI:")
	fmt.Printf("   redis-cli LLEN queue:%s\n", config.DefaultQueue)
	fmt.Printf("   redis-cli LRANGE queue:%s 0 -1\n", config.DefaultQueue)
	fmt.Printf("   redis-cli BRPOP queue:%s 0\n", config.DefaultQueue)
}