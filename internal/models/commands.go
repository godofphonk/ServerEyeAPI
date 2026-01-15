// Copyright (c) 2026 godofphonk
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package models

import "time"

// Command represents a command sent to a server
type Command struct {
	ID          string                 `json:"id" db:"id"`                     // Unique command ID
	Type        string                 `json:"type" db:"type"`                 // Command type (restart, update, etc.)
	Payload     map[string]interface{} `json:"payload" db:"payload"`           // Command parameters
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`     // When command was created
	Status      string                 `json:"status" db:"status"`             // pending, processed, failed
	ServerID    string                 `json:"server_id" db:"server_id"`       // Target server ID
	ProcessedAt *time.Time             `json:"processed_at" db:"processed_at"` // When command was processed
	Error       string                 `json:"error" db:"error"`               // Error message if failed
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
