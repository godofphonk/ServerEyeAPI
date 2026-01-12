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
