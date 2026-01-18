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

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ServersHandler handles server-related requests
type ServersHandler struct {
	storage storage.Storage
	logger  *logrus.Logger
}

// NewServersHandler creates a new servers handler
func NewServersHandler(storage storage.Storage, logger *logrus.Logger) *ServersHandler {
	return &ServersHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetServerByKey handles GET /api/servers/by-key/{server_key}
func (h *ServersHandler) GetServerByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		h.writeError(w, "server_key is required", http.StatusBadRequest)
		return
	}

	serverInfo, err := h.storage.GetServerByKey(r.Context(), serverKey)
	if err != nil {
		h.logger.WithError(err).WithField("server_key", serverKey).Error("Failed to get server by key")
		h.writeError(w, "Server not found", http.StatusNotFound)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"server_id": serverInfo.ServerID,
		"hostname":  serverInfo.Hostname,
	})
}

// ListServers handles GET /api/servers
func (h *ServersHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.storage.GetServers(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get servers")
		h.writeError(w, "Failed to get servers", http.StatusInternalServerError)
		return
	}

	// Get server details from storage
	serverDetails := make([]map[string]interface{}, 0)
	for _, serverInfo := range servers {
		status, err := h.storage.GetServerStatus(r.Context(), serverInfo.ServerID)
		if err != nil {
			h.logger.WithError(err).WithField("server_id", serverInfo.ServerID).Warn("Failed to get server status")
			status = &models.ServerStatus{Online: false}
		}

		serverDetails = append(serverDetails, map[string]interface{}{
			"server_id": serverInfo.ServerID,
			"status":    status,
		})
	}

	response := &models.ServerListResponse{
		Count:     len(serverDetails),
		Servers:   serverDetails,
		Timestamp: time.Now(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetServerStatus handles GET /api/servers/{server_id}/status
func (h *ServersHandler) GetServerStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		h.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	status, err := h.storage.GetServerStatus(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server status")
		h.writeError(w, "Failed to get server status", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"server_id": serverID,
		"status":    status,
	})
}

// writeJSON writes JSON response
func (h *ServersHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (h *ServersHandler) writeError(w http.ResponseWriter, message string, status int) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
