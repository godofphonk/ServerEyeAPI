package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
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
	var req models.SendCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ServerID == "" {
		h.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	if len(req.Command) == 0 {
		h.writeError(w, "command is required", http.StatusBadRequest)
		return
	}

	response, err := h.commandsService.SendCommand(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", req.ServerID).Error("Failed to send command")
		h.writeError(w, "Failed to send command", http.StatusInternalServerError)
		return
	}

	response.Timestamp = time.Now()
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
