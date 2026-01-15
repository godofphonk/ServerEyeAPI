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
	Sources      string    `json:"sources" db:"sources"` // TGBot, Web, TGBot,Web etc.
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
