package websocket

import (
	"encoding/json"
	"time"

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
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, logger *logrus.Logger) *Client {
	return &Client{
		conn:   conn,
		send:   make(chan models.WSMessage, 256),
		logger: logger,
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
	data, err := json.Marshal(msg)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal message")
		return false
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		c.logger.WithError(err).Error("Failed to send message")
		return false
	}

	return true
}

// Close closes the WebSocket connection
func (c *Client) Close() {
	c.conn.Close()
}

// Ping sends a ping message to the client
func (c *Client) Ping() error {
	return c.conn.WriteMessage(websocket.PingMessage, nil)
}
