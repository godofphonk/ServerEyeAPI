package models

import "time"

// DLQMessage represents a dead letter queue message
type DLQMessage struct {
	ID        string                 `json:"id" db:"id"`                 // Unique message ID
	Topic     string                 `json:"topic" db:"topic"`           // Kafka topic
	Message   []byte                 `json:"message" db:"message"`       // Original message
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`     // Message metadata
	Error     string                 `json:"error" db:"error"`           // Error message
	Status    string                 `json:"status" db:"status"`         // pending, processed, requeued, failed
	CreatedAt time.Time              `json:"created_at" db:"created_at"` // When message was added to DLQ
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"` // Last update time
}
