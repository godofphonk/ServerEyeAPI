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

	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/sirupsen/logrus"
)

// CommandsHandler handles command requests
type CommandsHandler struct {
	commandsService *services.CommandsService
	logger          *logrus.Logger
}

// NewCommandsHandler creates a new commands handler
func NewCommandsHandler(commandsService *services.CommandsService, logger *logrus.Logger) *CommandsHandler {
	return &CommandsHandler{
		commandsService: commandsService,
		logger:          logger,
	}
}

// SendCommand handles POST /api/servers/{server_id}/command
func (h *CommandsHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	var req services.SendCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ServerID == "" {
		h.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		h.writeError(w, "command type is required", http.StatusBadRequest)
		return
	}

	response, err := h.commandsService.SendCommand(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", req.ServerID).Error("Failed to send command")
		h.writeError(w, "Failed to send command", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes JSON response
func (h *CommandsHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (h *CommandsHandler) writeError(w http.ResponseWriter, message string, status int) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
