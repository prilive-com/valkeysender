package valkeysender

import (
	"reflect"
	"testing"
	"time"
)

func TestJSONSerializer(t *testing.T) {
	serializer := NewJSONSerializer()
	
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string message",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "byte slice message",
			input:    []byte("hello bytes"),
			expected: "hello bytes",
		},
		{
			name:     "simple struct",
			input:    struct{ Name string }{Name: "test"},
			expected: `{"Name":"test"}`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test serialization
			data, err := serializer.Serialize(tt.input)
			if err != nil {
				t.Fatalf("Serialize failed: %v", err)
			}
			
			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
			
			// Test content type
			if serializer.ContentType() != "application/json" {
				t.Errorf("Expected content type application/json, got %s", serializer.ContentType())
			}
		})
	}
}

func TestJSONSerializerDeserialization(t *testing.T) {
	serializer := NewJSONSerializer()
	
	t.Run("deserialize to string", func(t *testing.T) {
		data := []byte("hello world")
		var result string
		
		err := serializer.Deserialize(data, &result)
		if err != nil {
			t.Fatalf("Deserialize failed: %v", err)
		}
		
		if result != "hello world" {
			t.Errorf("Expected 'hello world', got %s", result)
		}
	})
	
	t.Run("deserialize to byte slice", func(t *testing.T) {
		data := []byte("hello bytes")
		var result []byte
		
		err := serializer.Deserialize(data, &result)
		if err != nil {
			t.Fatalf("Deserialize failed: %v", err)
		}
		
		if !reflect.DeepEqual(result, data) {
			t.Errorf("Expected %v, got %v", data, result)
		}
	})
	
	t.Run("deserialize JSON struct", func(t *testing.T) {
		data := []byte(`{"Name":"test","Value":42}`)
		var result struct {
			Name  string
			Value int
		}
		
		err := serializer.Deserialize(data, &result)
		if err != nil {
			t.Fatalf("Deserialize failed: %v", err)
		}
		
		if result.Name != "test" || result.Value != 42 {
			t.Errorf("Expected {Name:test Value:42}, got %+v", result)
		}
	})
}

func TestUserRegistrationSerialization(t *testing.T) {
	userData := UserRegistrationData{
		Name:             "John Doe",
		Email:            "john@example.com",
		TelegramUserID:   123456789,
		TelegramUsername: "johndoe",
		FirstName:        "John",
		LastName:         "Doe",
		PhoneNumber:      "+1234567890",
		LanguageCode:     "en",
		RegistrationTime: time.Date(2025, 5, 28, 12, 0, 0, 0, time.UTC),
		Source:           "test",
	}
	
	// Test serialization
	data, err := SerializeUserRegistration(userData)
	if err != nil {
		t.Fatalf("SerializeUserRegistration failed: %v", err)
	}
	
	// Test deserialization
	result, err := DeserializeUserRegistration(data)
	if err != nil {
		t.Fatalf("DeserializeUserRegistration failed: %v", err)
	}
	
	// Compare important fields
	if result.Name != userData.Name {
		t.Errorf("Expected name %s, got %s", userData.Name, result.Name)
	}
	if result.Email != userData.Email {
		t.Errorf("Expected email %s, got %s", userData.Email, result.Email)
	}
	if result.TelegramUserID != userData.TelegramUserID {
		t.Errorf("Expected telegram user ID %d, got %d", userData.TelegramUserID, result.TelegramUserID)
	}
	if result.Source != userData.Source {
		t.Errorf("Expected source %s, got %s", userData.Source, result.Source)
	}
}

func TestMessageEnvelopeSerialization(t *testing.T) {
	envelope := MessageEnvelope{
		ID:        "test-id-123",
		Queue:     "test-queue",
		Payload:   []byte("test payload"),
		Headers:   map[string]string{"content-type": "text/plain"},
		Timestamp: time.Date(2025, 5, 28, 12, 0, 0, 0, time.UTC),
		TTL:       24 * time.Hour,
		Retries:   0,
		Metadata:  map[string]interface{}{"test": "value"},
	}
	
	// Test serialization
	data, err := SerializeMessageEnvelope(envelope)
	if err != nil {
		t.Fatalf("SerializeMessageEnvelope failed: %v", err)
	}
	
	// Test deserialization
	result, err := DeserializeMessageEnvelope(data)
	if err != nil {
		t.Fatalf("DeserializeMessageEnvelope failed: %v", err)
	}
	
	// Compare important fields
	if result.ID != envelope.ID {
		t.Errorf("Expected ID %s, got %s", envelope.ID, result.ID)
	}
	if result.Queue != envelope.Queue {
		t.Errorf("Expected queue %s, got %s", envelope.Queue, result.Queue)
	}
	if !reflect.DeepEqual(result.Payload, envelope.Payload) {
		t.Errorf("Expected payload %v, got %v", envelope.Payload, result.Payload)
	}
	if result.TTL != envelope.TTL {
		t.Errorf("Expected TTL %v, got %v", envelope.TTL, result.TTL)
	}
}

func TestSerializerErrorCases(t *testing.T) {
	serializer := NewJSONSerializer()
	
	t.Run("serialize nil", func(t *testing.T) {
		_, err := serializer.Serialize(nil)
		if err == nil {
			t.Error("Expected error for nil input")
		}
	})
	
	t.Run("deserialize empty data", func(t *testing.T) {
		var result string
		err := serializer.Deserialize([]byte{}, &result)
		if err == nil {
			t.Error("Expected error for empty data")
		}
	})
	
	t.Run("deserialize to nil target", func(t *testing.T) {
		err := serializer.Deserialize([]byte("test"), nil)
		if err == nil {
			t.Error("Expected error for nil target")
		}
	})
	
	t.Run("deserialize invalid JSON", func(t *testing.T) {
		var result struct{ Name string }
		err := serializer.Deserialize([]byte("{invalid json"), &result)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}