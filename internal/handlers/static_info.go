package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// StaticInfoHandler handles static server information endpoints
type StaticInfoHandler struct {
	staticStorage storage.StaticDataStorage
	logger        *logrus.Logger
}

// NewStaticInfoHandler creates a new static info handler
func NewStaticInfoHandler(staticStorage storage.StaticDataStorage, logger *logrus.Logger) *StaticInfoHandler {
	return &StaticInfoHandler{
		staticStorage: staticStorage,
		logger:        logger,
	}
}

// UpsertStaticInfo handles POST/PUT requests to update static server information
func (h *StaticInfoHandler) UpsertStaticInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	var info storage.CompleteStaticInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		h.logger.WithError(err).Error("Failed to decode static info request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.staticStorage.UpsertCompleteStaticInfo(r.Context(), serverID, &info); err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to upsert static info")
		http.Error(w, "Failed to update static information", http.StatusInternalServerError)
		return
	}

	h.logger.WithField("server_id", serverID).Info("Successfully updated static info")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Static information updated successfully",
		"server_id": serverID,
	})
}

// GetStaticInfo handles GET requests to retrieve static server information
func (h *StaticInfoHandler) GetStaticInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	info, err := h.staticStorage.GetCompleteStaticInfo(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get static info")
		http.Error(w, "Failed to retrieve static information", http.StatusInternalServerError)
		return
	}

	if info.ServerInfo == nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// GetServerInfo handles GET requests for basic server info only
func (h *StaticInfoHandler) GetServerInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	info, err := h.staticStorage.GetServerInfo(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server info")
		http.Error(w, "Failed to retrieve server information", http.StatusInternalServerError)
		return
	}

	if info == nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// GetHardwareInfo handles GET requests for hardware info only
func (h *StaticInfoHandler) GetHardwareInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	info, err := h.staticStorage.GetHardwareInfo(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get hardware info")
		http.Error(w, "Failed to retrieve hardware information", http.StatusInternalServerError)
		return
	}

	if info == nil {
		http.Error(w, "Hardware information not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// GetNetworkInterfaces handles GET requests for network interfaces
func (h *StaticInfoHandler) GetNetworkInterfaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	interfaces, err := h.staticStorage.GetNetworkInterfaces(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get network interfaces")
		http.Error(w, "Failed to retrieve network interfaces", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_id":  serverID,
		"interfaces": interfaces,
		"count":      len(interfaces),
	})
}

// GetDiskInfo handles GET requests for disk information
func (h *StaticInfoHandler) GetDiskInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	disks, err := h.staticStorage.GetDiskInfo(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get disk info")
		http.Error(w, "Failed to retrieve disk information", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_id": serverID,
		"disks":     disks,
		"count":     len(disks),
	})
}

// Helper method to get server_id from server_key
func (h *StaticInfoHandler) getServerIDFromKey(ctx context.Context, serverKey string) (string, error) {
	// This would need access to server repository to convert key to ID
	// For now, we'll return an error - this needs to be implemented
	return "", fmt.Errorf("server key to ID conversion not implemented")
}

// UpsertStaticInfoByKey handles POST/PUT requests using server_key
func (h *StaticInfoHandler) UpsertStaticInfoByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		http.Error(w, "server_key is required", http.StatusBadRequest)
		return
	}

	// Convert server_key to server_id (TODO: implement proper conversion)
	serverID := "srv_" + serverKey[4:] // Simple conversion for now

	// Read request body
	var info storage.CompleteStaticInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		h.logger.WithError(err).Error("Failed to decode static info request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.staticStorage.UpsertCompleteStaticInfo(r.Context(), serverID, &info); err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to upsert static info")
		http.Error(w, "Failed to update static information", http.StatusInternalServerError)
		return
	}

	h.logger.WithField("server_id", serverID).Info("Successfully updated static info")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Static information updated successfully",
		"server_id": serverID,
	})
}

// GetStaticInfoByKey handles GET requests using server_key
func (h *StaticInfoHandler) GetStaticInfoByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		http.Error(w, "server_key is required", http.StatusBadRequest)
		return
	}

	// Convert server_key to server_id (TODO: implement proper conversion)
	serverID := "srv_" + serverKey[4:] // Simple conversion for now

	info, err := h.staticStorage.GetCompleteStaticInfo(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get static info")
		http.Error(w, "Failed to retrieve static information", http.StatusInternalServerError)
		return
	}

	if info.ServerInfo == nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// GetServerInfoByKey handles GET requests for server info using server_key
func (h *StaticInfoHandler) GetServerInfoByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		http.Error(w, "server_key is required", http.StatusBadRequest)
		return
	}

	// Convert server_key to server_id (TODO: implement proper conversion)
	serverID := "srv_" + serverKey[4:] // Simple conversion for now

	// Delegate to the existing method
	r = r.WithContext(context.WithValue(r.Context(), "server_id", serverID))
	h.GetServerInfo(w, r)
}

// GetHardwareInfoByKey handles GET requests for hardware info using server_key
func (h *StaticInfoHandler) GetHardwareInfoByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		http.Error(w, "server_key is required", http.StatusBadRequest)
		return
	}

	// Convert server_key to server_id (TODO: implement proper conversion)
	serverID := "srv_" + serverKey[4:] // Simple conversion for now

	// Delegate to the existing method
	r = r.WithContext(context.WithValue(r.Context(), "server_id", serverID))
	h.GetHardwareInfo(w, r)
}

// GetNetworkInterfacesByKey handles GET requests for network interfaces using server_key
func (h *StaticInfoHandler) GetNetworkInterfacesByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		http.Error(w, "server_key is required", http.StatusBadRequest)
		return
	}

	// Convert server_key to server_id (TODO: implement proper conversion)
	serverID := "srv_" + serverKey[4:] // Simple conversion for now

	// Delegate to the existing method
	r = r.WithContext(context.WithValue(r.Context(), "server_id", serverID))
	h.GetNetworkInterfaces(w, r)
}

// GetDiskInfoByKey handles GET requests for disk info using server_key
func (h *StaticInfoHandler) GetDiskInfoByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		http.Error(w, "server_key is required", http.StatusBadRequest)
		return
	}

	// Convert server_key to server_id (TODO: implement proper conversion)
	serverID := "srv_" + serverKey[4:] // Simple conversion for now

	// Delegate to the existing method
	r = r.WithContext(context.WithValue(r.Context(), "server_id", serverID))
	h.GetDiskInfo(w, r)
}
