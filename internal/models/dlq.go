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
