package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/godofphonk/ServerEyeAPI/pkg/models"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WSClient struct {
	conn      *websocket.Conn
	serverID  string
	serverKey string
	isAgent   bool
	send      chan WSMessage
	logger    *logrus.Logger
	mu        sync.Mutex
}

type WebSocketServer struct {
	server     *Server
	clients    map[*WSClient]bool
	clientsM   sync.RWMutex
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan WSMessage
	logger     *logrus.Logger
}

type WSMessage struct {
	Type      string                 `json:"type"`
	ServerID  string                 `json:"server_id,omitempty"`
	ServerKey string                 `json:"server_key,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type WSSubscribeMessage struct {
	ServerID string `json:"server_id"`
	Metric   string `json:"metric"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

func NewWebSocketServer(server *Server, logger *logrus.Logger) *WebSocketServer {
	wss := &WebSocketServer{
		server:     server,
		clients:    make(map[*WSClient]bool),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan WSMessage),
		logger:     logger,
	}

	// Start the WebSocket server loop
	go wss.run()

	return wss
}

func (wss *WebSocketServer) run() {
	for {
		select {
		case client := <-wss.register:
			wss.clientsM.Lock()
			wss.clients[client] = true
			wss.clientsM.Unlock()
			wss.logger.Info("WebSocket client connected")

		case client := <-wss.unregister:
			wss.clientsM.Lock()
			if _, ok := wss.clients[client]; ok {
				delete(wss.clients, client)
				wss.logger.WithField("server_id", client.serverID).Info("WebSocket client disconnected")
			}
			wss.clientsM.Unlock()

		case message := <-wss.broadcast:
			wss.clientsM.RLock()
			for client := range wss.clients {
				// Don't send metrics back to the agent that sent them
				if message.Type == "metrics" && client.isAgent && client.serverID == message.ServerID {
					continue
				}

				select {
				case client.send <- message:
				default:
					delete(wss.clients, client)
				}
			}
			wss.clientsM.RUnlock()
		}
	}
}

func (wss *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wss.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}

	client := &WSClient{
		conn:   conn,
		send:   make(chan WSMessage, 256),
		logger: wss.logger,
	}

	wss.register <- client

	go client.writePump()
	go client.readPump(wss)
}

func (c *WSClient) readPump(wss *WebSocketServer) {
	defer func() {
		wss.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg WSMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.WithError(err).Error("WebSocket error")
			}
			break
		}

		msg.Timestamp = time.Now()

		switch msg.Type {
		case "auth":
			// Authenticate agent
			if msg.ServerID != "" && msg.ServerKey != "" {
				c.mu.Lock()
				c.serverID = msg.ServerID
				c.serverKey = msg.ServerKey
				c.isAgent = true
				c.mu.Unlock()

				c.send <- WSMessage{
					Type:      "auth_success",
					Timestamp: time.Now(),
				}
				c.logger.WithField("server_id", msg.ServerID).Info("Agent authenticated")

				// Broadcast that server is online
				wss.broadcast <- WSMessage{
					Type:      "server_online",
					ServerID:  msg.ServerID,
					Timestamp: time.Now(),
				}
			} else {
				c.send <- WSMessage{
					Type:      "auth_error",
					Data:      map[string]interface{}{"error": "server_id and server_key required"},
					Timestamp: time.Now(),
				}
			}

		case "metrics":
			// Store metrics and broadcast
			if c.isAgent {
				// Store in database
				metric := &models.Metric{
					ServerID:  msg.ServerID,
					ServerKey: msg.ServerKey,
					Type:      "system",
					Value:     0, // We'll handle individual metrics in Data
					Timestamp: msg.Timestamp,
					Tags:      make(map[string]string),
				}

				// Store metric
				err := wss.server.storage.StoreMetric(context.Background(), metric)
				if err != nil {
					c.logger.WithError(err).Error("Failed to store metric")
				}

				// Broadcast to all dashboard clients
				wss.broadcast <- msg
			}

		case "heartbeat":
			// Agent heartbeat
			if c.isAgent {
				wss.broadcast <- WSMessage{
					Type:      "server_online",
					ServerID:  msg.ServerID,
					Timestamp: time.Now(),
				}
			}

		case "subscribe":
			// Dashboard client wants to subscribe to metrics
			if !c.isAgent {
				wss.handleSubscription(c, msg)
			}
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				c.logger.WithError(err).Error("Failed to write message")
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (wss *WebSocketServer) handleSubscription(client *WSClient, msg WSMessage) {
	data, err := json.Marshal(msg.Data)
	if err != nil {
		wss.logger.WithError(err).Error("Failed to marshal subscription data")
		return
	}

	var sub WSSubscribeMessage
	if err := json.Unmarshal(data, &sub); err != nil {
		wss.logger.WithError(err).Error("Failed to unmarshal subscription message")
		return
	}

	// Send current metrics for the subscription
	if sub.ServerID != "" {
		metrics, err := wss.server.storage.GetLatestMetrics(context.Background(), sub.ServerID)
		if err != nil {
			wss.logger.WithError(err).WithField("server", sub.ServerID).Error("Failed to get latest metrics")
			return
		}

		response := WSMessage{
			Type:      "metrics",
			ServerID:  sub.ServerID,
			Data:      map[string]interface{}{"metrics": metrics},
			Timestamp: time.Now(),
		}

		select {
		case client.send <- response:
		default:
			wss.logger.Warn("Failed to send metrics to client - channel full")
		}
	}
}

// BroadcastMetric sends metric to all connected clients
func (wss *WebSocketServer) BroadcastMetric(serverID string, data map[string]interface{}) {
	message := WSMessage{
		Type:      "metrics",
		ServerID:  serverID,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case wss.broadcast <- message:
	default:
		wss.logger.Warn("Broadcast channel full")
	}
}

// GetClientCount returns the number of connected WebSocket clients
func (wss *WebSocketServer) GetClientCount() int {
	wss.clientsM.RLock()
	defer wss.clientsM.RUnlock()
	return len(wss.clients)
}
