package models

import "time"

// Connection represents a connection entity
type Connection struct {
	ID             string                 `json:"id" db:"id"`                           // Unique connection ID
	ServerID       string                 `json:"server_id" db:"server_id"`             // Server ID
	Type           string                 `json:"type" db:"type"`                       // Connection type (websocket, http)
	RemoteAddr     string                 `json:"remote_addr" db:"remote_addr"`         // Remote address
	UserAgent      string                 `json:"user_agent" db:"user_agent"`           // User agent string
	Status         string                 `json:"status" db:"status"`                   // Connection status (active, disconnected)
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`               // Additional metadata
	ConnectedAt    time.Time              `json:"connected_at" db:"connected_at"`       // When connection was established
	DisconnectedAt *time.Time             `json:"disconnected_at" db:"disconnected_at"` // When connection was closed
	LastActivity   time.Time              `json:"last_activity" db:"last_activity"`     // Last activity timestamp
}
