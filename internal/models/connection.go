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
