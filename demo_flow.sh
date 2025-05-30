#!/bin/bash

# Demo script showing complete message flow with valkeysender
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${PURPLE}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║ $1${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════════════════════╝${NC}\n"
}

print_step() {
    echo -e "${CYAN}🔹 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

main() {
    print_header "🚀 VALKEY SENDER COMPLETE MESSAGE FLOW DEMO"
    
    echo -e "${BLUE}This demo will show you:${NC}"
    echo -e "${BLUE}  1. 📤 Send messages to Valkey using valkeysender${NC}"
    echo -e "${BLUE}  2. 🔍 Inspect messages in Valkey queues${NC}"
    echo -e "${BLUE}  3. 📥 Consume messages using Redis CLI${NC}"
    echo -e "${BLUE}  4. 📊 Monitor queue statistics${NC}"
    echo
    
    read -p "Press Enter to start the demo..."
    
    # Step 1: Check if Redis/Valkey is running
    print_header "📋 STEP 1: Environment Check"
    
    print_step "Checking if Redis/Valkey is running on localhost:6379..."
    if redis-cli ping > /dev/null 2>&1; then
        print_success "Redis/Valkey is running and responding"
    else
        print_warning "Redis/Valkey is not running on localhost:6379"
        echo -e "${YELLOW}Please start Redis/Valkey before continuing:${NC}"
        echo -e "${BLUE}  • Using Docker: docker run -d -p 6379:6379 redis:alpine${NC}"
        echo -e "${BLUE}  • Using local install: redis-server${NC}"
        echo -e "${BLUE}  • Using Valkey: valkey-server${NC}"
        echo
        read -p "Press Enter when Redis/Valkey is running..."
        
        if ! redis-cli ping > /dev/null 2>&1; then
            echo -e "${RED}❌ Still cannot connect to Redis/Valkey${NC}"
            exit 1
        fi
        print_success "Redis/Valkey is now running"
    fi
    
    # Step 2: Clear any existing test data
    print_header "🧹 STEP 2: Cleaning Up Previous Test Data"
    
    print_step "Removing any existing test queues..."
    redis-cli DEL queue:user-registrations queue:test-queue queue:temporary-messages queue:batch-messages > /dev/null 2>&1 || true
    print_success "Test environment cleaned"
    
    # Step 3: Build and run the sender
    print_header "📤 STEP 3: Sending Messages with valkeysender"
    
    print_step "Building valkeysender test application..."
    cd /home/toha/telegram/valkeysender
    if go build -o test-simple test_simple.go; then
        print_success "Test application built successfully"
    else
        echo -e "${RED}❌ Failed to build test application${NC}"
        exit 1
    fi
    
    print_step "Running simple test to send messages..."
    echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════════════════${NC}"
    ./test-simple
    echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════════════════${NC}"
    print_success "Messages sent to Valkey!"
    
    # Step 4: Inspect queues
    print_header "🔍 STEP 4: Inspecting Valkey Queues"
    
    print_step "Checking queue sizes..."
    test_queue_size=$(redis-cli LLEN queue:test-queue)
    echo -e "${BLUE}📊 Queue 'test-queue': ${test_queue_size} messages${NC}"
    
    if [ "$test_queue_size" -gt 0 ]; then
        print_step "Inspecting messages in test-queue (without consuming)..."
        echo -e "${GREEN}Messages in queue:${NC}"
        redis-cli LRANGE queue:test-queue 0 -1 | head -5
        echo
    fi
    
    # Step 5: Consume messages
    print_header "📥 STEP 5: Consuming Messages"
    
    print_step "Consuming messages from test-queue using Redis CLI..."
    echo -e "${BLUE}We'll use BRPOP to consume messages (blocks until message available)${NC}"
    echo
    
    while [ "$(redis-cli LLEN queue:test-queue)" -gt 0 ]; do
        echo -e "${CYAN}📨 Consuming next message...${NC}"
        message=$(redis-cli BRPOP queue:test-queue 1)
        if [ -n "$message" ]; then
            echo -e "${GREEN}✅ Consumed: ${message}${NC}"
        fi
        echo
    done
    
    print_success "All messages consumed from test-queue!"
    
    # Step 6: Advanced demo
    print_header "🚀 STEP 6: Advanced Demo with Multiple Message Types"
    
    print_step "Running full demo application..."
    echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════════════════${NC}"
    
    # Create minimal config for demo
    export VALKEY_SENDER_ADDRESS=localhost:6379
    export VALKEY_SENDER_LOG_LEVEL=INFO
    
    if go build -o demo-sender ./example/; then
        timeout 30s ./demo-sender || true
    else
        print_warning "Could not build full demo application"
    fi
    
    echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════════════════${NC}"
    
    # Step 7: Show final queue states
    print_header "📊 STEP 7: Final Queue Statistics"
    
    print_step "Checking all queue sizes..."
    for queue in user-registrations temporary-messages batch-messages; do
        size=$(redis-cli LLEN queue:$queue 2>/dev/null || echo "0")
        echo -e "${BLUE}📋 Queue '$queue': ${size} messages${NC}"
        
        if [ "$size" -gt 0 ]; then
            echo -e "${GREEN}   Sample message:${NC}"
            redis-cli LINDEX queue:$queue 0 2>/dev/null | head -c 100
            echo "..."
            echo
        fi
    done
    
    # Step 8: Show consumption commands
    print_header "🛠️  STEP 8: How to Consume Messages"
    
    print_info "You can consume messages using Redis CLI:"
    echo -e "${CYAN}  # Blocking pop (waits for messages):${NC}"
    echo -e "${BLUE}  redis-cli BRPOP queue:user-registrations 0${NC}"
    echo
    echo -e "${CYAN}  # Non-blocking pop:${NC}"
    echo -e "${BLUE}  redis-cli RPOP queue:user-registrations${NC}"
    echo
    echo -e "${CYAN}  # Inspect without consuming:${NC}"
    echo -e "${BLUE}  redis-cli LRANGE queue:user-registrations 0 -1${NC}"
    echo
    echo -e "${CYAN}  # Monitor queue size:${NC}"
    echo -e "${BLUE}  redis-cli LLEN queue:user-registrations${NC}"
    echo
    
    # Final summary
    print_header "🎉 DEMO COMPLETED SUCCESSFULLY!"
    
    print_success "valkeysender is working perfectly!"
    echo
    print_info "Summary of what happened:"
    echo -e "${GREEN}  1. ✅ Connected to Valkey/Redis successfully${NC}"
    echo -e "${GREEN}  2. ✅ Sent messages using valkeysender library${NC}"
    echo -e "${GREEN}  3. ✅ Messages stored in Redis Lists (queue:*)${NC}"
    echo -e "${GREEN}  4. ✅ Consumed messages using Redis CLI${NC}"
    echo -e "${GREEN}  5. ✅ Demonstrated complete message flow!${NC}"
    echo
    
    print_info "Next steps for production:"
    echo -e "${BLUE}  🔹 Replace kafkasender with valkeysender in your dispatcher${NC}"
    echo -e "${BLUE}  🔹 Create valkeyreceiver library for consuming services${NC}"
    echo -e "${BLUE}  🔹 No complex Kafka authentication needed!${NC}"
    echo -e "${BLUE}  🔹 Much simpler operations and debugging${NC}"
    echo
    
    print_success "Thank you for trying valkeysender! 🚀"
}

# Check if redis-cli is available
if ! command -v redis-cli &> /dev/null; then
    echo -e "${RED}❌ redis-cli is not available${NC}"
    echo -e "${YELLOW}Please install Redis CLI:${NC}"
    echo -e "${BLUE}  • Ubuntu/Debian: apt-get install redis-tools${NC}"
    echo -e "${BLUE}  • CentOS/RHEL: yum install redis${NC}"
    echo -e "${BLUE}  • macOS: brew install redis${NC}"
    exit 1
fi

# Run the main demo
main "$@"