package models

import "time"

// Command represents a command sent to a server
type Command struct {
	ID        string                 `json:"id"`         // Unique command ID
	Type      string                 `json:"type"`       // Command type (restart, update, etc.)
	Payload   map[string]interface{} `json:"payload"`    // Command parameters
	CreatedAt time.Time              `json:"created_at"` // When command was created
	Status    string                 `json:"status"`     // pending, executed, failed
	ServerID  string                 `json:"server_id"`  // Target server ID
}

// CommandPayload represents different command types
type RestartCommand struct {
	Force bool `json:"force"`
}

type UpdateCommand struct {
	Version string `json:"version"`
	Force   bool   `json:"force"`
}

type ScriptCommand struct {
	Script string   `json:"script"`
	Args   []string `json:"args"`
}

// CommandExecutionResult represents the result of command execution
type CommandExecutionResult struct {
	CommandID  string    `json:"command_id"`
	ServerID   string    `json:"server_id"`
	Status     string    `json:"status"`
	Output     string    `json:"output"`
	Error      string    `json:"error"`
	ExecutedAt time.Time `json:"executed_at"`
}
