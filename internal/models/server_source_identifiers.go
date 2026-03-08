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

// ServerSourceIdentifier represents an identifier for a server source
type ServerSourceIdentifier struct {
	ID             int64                  `json:"id" db:"id"`
	ServerID       string                 `json:"server_id" db:"server_id"`
	SourceType     string                 `json:"source_type" db:"source_type"`           // TGBot, Web, Email, etc.
	Identifier     string                 `json:"identifier" db:"identifier"`             // TG ID, user ID, email
	IdentifierType string                 `json:"identifier_type" db:"identifier_type"`   // telegram_id, user_id, email
	TelegramID     *int64                 `json:"telegram_id,omitempty" db:"telegram_id"` // Optional Telegram ID for account linking
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`                 // additional info
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// SourceIdentifierRequest represents a request to add/update source identifiers
type SourceIdentifierRequest struct {
	SourceType     string                 `json:"source_type" validate:"required,oneof=TGBot Web Email"`               // TGBot, Web, Email
	Identifiers    []string               `json:"identifiers" validate:"required,min=1"`                               // TG IDs, user IDs, emails
	IdentifierType string                 `json:"identifier_type" validate:"required,oneof=telegram_id user_id email"` // telegram_id, user_id, email
	TelegramID     *int64                 `json:"telegram_id,omitempty"`                                               // Optional Telegram ID for account linking
	Metadata       map[string]interface{} `json:"metadata,omitempty"`                                                  // optional metadata
}

// SourceIdentifierResponse represents response with server source identifiers
type SourceIdentifierResponse struct {
	ServerID    string                   `json:"server_id"`
	SourceType  string                   `json:"source_type"`
	Identifiers []ServerSourceIdentifier `json:"identifiers"`
}

// ServerSourcesResponse represents combined response with all server sources and identifiers
type ServerSourcesResponse struct {
	ServerID    string                              `json:"server_id"`
	Sources     []string                            `json:"sources"`     // TGBot, Web (legacy)
	Identifiers map[string][]ServerSourceIdentifier `json:"identifiers"` // source_type -> identifiers
}
