package models

// WSMessage represents WebSocket message structure
type WSMessage struct {
	Type      string                 `json:"type"`
	ServerID  string                 `json:"server_id,omitempty"`
	ServerKey string                 `json:"server_key,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp,omitempty"`
}

// WSClient represents a WebSocket client
type WSClient struct {
	ID        string
	ServerID  string
	ServerKey string
	IsAgent   bool
	Send      chan WSMessage
}

// WSMessageType represents WebSocket message types
const (
	WSMessageTypeAuth        = "auth"
	WSMessageTypeAuthSuccess = "auth_success"
	WSMessageTypeError       = "error"
	WSMessageTypeMetrics     = "metrics"
	WSMessageTypeHeartbeat   = "heartbeat"
	WSMessageTypeCommand     = "command"
	WSMessageTypeSubscribe   = "subscribe"
)
