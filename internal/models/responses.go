package models

import "time"

// RegisterKeyResponse represents response after successful registration
type RegisterKeyResponse struct {
	ServerID  string `json:"server_id"`
	ServerKey string `json:"server_key"`
	Status    string `json:"status"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Clients   int       `json:"clients,omitempty"`
}

// ServerListResponse represents list of servers response
type ServerListResponse struct {
	Count     int                      `json:"count"`
	Servers   []map[string]interface{} `json:"servers"`
	Timestamp time.Time                `json:"timestamp"`
}

// CommandResponse represents command sending response
type CommandResponse struct {
	Message   string                 `json:"message"`
	ServerID  string                 `json:"server_id"`
	Command   map[string]interface{} `json:"command"`
	Timestamp time.Time              `json:"timestamp"`
}
