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

package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Client represents a WebSocket client
type Client struct {
	conn      *websocket.Conn
	ServerID  string
	ServerKey string
	IsAgent   bool
	send      chan models.WSMessage
	logger    *logrus.Logger
	config    *config.Config
	closed    bool
	mutex     sync.RWMutex
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, logger *logrus.Logger, cfg *config.Config) *Client {
	// Set initial read deadline
	conn.SetReadDeadline(time.Now().Add(cfg.WebSocket.PongWait))
	conn.SetWriteDeadline(time.Now().Add(cfg.WebSocket.WriteTimeout))

	// Set pong handler to extend read deadline on pong responses
	conn.SetPongHandler(func(appData string) error {
		logger.Debug("Received pong from client, extending read deadline")
		return conn.SetReadDeadline(time.Now().Add(cfg.WebSocket.PongWait))
	})

	return &Client{
		conn:   conn,
		send:   make(chan models.WSMessage, cfg.WebSocket.BufferSize),
		logger: logger,
		config: cfg,
		closed: false,
	}
}

// ReadMessage reads a message from the WebSocket connection
func (c *Client) ReadMessage() (models.WSMessage, error) {
	var msg models.WSMessage

	// Read message without setting deadline here
	// Deadline is set by pong handler and main server loop
	err := c.conn.ReadJSON(&msg)
	return msg, err
}

// SendMessage sends a message to the WebSocket connection
func (c *Client) SendMessage(msg models.WSMessage) bool {
	c.mutex.RLock()
	if c.closed {
		c.mutex.RUnlock()
		return false
	}
	c.mutex.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal message")
		return false
	}

	c.conn.SetWriteDeadline(time.Now().Add(c.config.WebSocket.WriteTimeout))
	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		c.logger.WithError(err).Error("Failed to send message")
		return false
	}

	return true
}

// Close closes the WebSocket connection safely
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return // Already closed
	}

	c.closed = true

	// Close the send channel
	if c.send != nil {
		close(c.send)
		c.send = nil
	}

	// Close the connection
	if c.conn != nil {
		c.conn.Close()
	}
}

// IsClosed returns true if the client is closed
func (c *Client) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.closed
}

// Ping sends a ping message to the client
func (c *Client) Ping() error {
	c.mutex.RLock()
	if c.closed {
		c.mutex.RUnlock()
		return nil // Already closed, ignore
	}
	c.mutex.RUnlock()

	return c.conn.WriteMessage(websocket.PingMessage, nil)
}
