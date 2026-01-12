package models

import "time"

// ServerInfo represents server information for authentication
type ServerInfo struct {
	ServerID  string `json:"server_id"`
	SecretKey string `json:"secret_key"`
	Hostname  string `json:"hostname"`
}

// Server represents a server entity
type Server struct {
	ID           string    `json:"id" db:"server_id"`
	ServerKey    string    `json:"server_key" db:"server_key"`
	SecretKey    string    `json:"secret_key" db:"secret_key"`
	Hostname     string    `json:"hostname" db:"hostname"`
	OSInfo       string    `json:"os_info" db:"os_info"`
	AgentVersion string    `json:"agent_version" db:"agent_version"`
	Status       string    `json:"status" db:"status"`
	LastSeen     time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// GeneratedKey represents a generated key entity
type GeneratedKey struct {
	ID           int64     `json:"id" db:"id"`
	ServerID     string    `json:"server_id" db:"server_id"`
	ServerKey    string    `json:"server_key" db:"server_key"`
	AgentVersion string    `json:"agent_version" db:"agent_version"`
	OSInfo       string    `json:"os_info" db:"os_info"`
	Hostname     string    `json:"hostname" db:"hostname"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
