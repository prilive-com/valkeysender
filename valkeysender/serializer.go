package valkeysender

import (
	"encoding/json"
	"fmt"
)

// JSONSerializer implements MessageSerializer using JSON encoding
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Serialize converts a message to JSON bytes
func (s *JSONSerializer) Serialize(message interface{}) ([]byte, error) {
	if message == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}
	
	// Handle string messages directly
	if str, ok := message.(string); ok {
		return []byte(str), nil
	}
	
	// Handle byte slice messages directly
	if bytes, ok := message.([]byte); ok {
		return bytes, nil
	}
	
	// Serialize other types as JSON
	data, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message to JSON: %w", err)
	}
	
	return data, nil
}

// Deserialize converts JSON bytes back to a message
func (s *JSONSerializer) Deserialize(data []byte, target interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}
	
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}
	
	// Handle string target
	if str, ok := target.(*string); ok {
		*str = string(data)
		return nil
	}
	
	// Handle byte slice target
	if bytes, ok := target.(*[]byte); ok {
		*bytes = make([]byte, len(data))
		copy(*bytes, data)
		return nil
	}
	
	// Deserialize JSON to target
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to deserialize JSON to target: %w", err)
	}
	
	return nil
}

// ContentType returns the content type for JSON
func (s *JSONSerializer) ContentType() string {
	return "application/json"
}

// SerializeUserRegistration is a convenience method for user registration data
func SerializeUserRegistration(userData UserRegistrationData) ([]byte, error) {
	serializer := NewJSONSerializer()
	return serializer.Serialize(userData)
}

// DeserializeUserRegistration is a convenience method for user registration data
func DeserializeUserRegistration(data []byte) (UserRegistrationData, error) {
	var userData UserRegistrationData
	serializer := NewJSONSerializer()
	err := serializer.Deserialize(data, &userData)
	return userData, err
}

// SerializeMessageEnvelope serializes a message envelope
func SerializeMessageEnvelope(envelope MessageEnvelope) ([]byte, error) {
	serializer := NewJSONSerializer()
	return serializer.Serialize(envelope)
}

// DeserializeMessageEnvelope deserializes a message envelope
func DeserializeMessageEnvelope(data []byte) (MessageEnvelope, error) {
	var envelope MessageEnvelope
	serializer := NewJSONSerializer()
	err := serializer.Deserialize(data, &envelope)
	return envelope, err
}