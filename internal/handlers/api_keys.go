package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/godofphonk/ServerEyeAPI/internal/storage"
)

type APIKeyHandler struct {
	storage *storage.APIKeyStorage
	logger  *logrus.Logger
}

func NewAPIKeyHandler(storage *storage.APIKeyStorage, logger *logrus.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		storage: storage,
		logger:  logger,
	}
}

type CreateAPIKeyRequest struct {
	ServiceID   string   `json:"service_id"`
	ServiceName string   `json:"service_name"`
	Permissions []string `json:"permissions"`
	ExpiresIn   string   `json:"expires_in,omitempty"` // "30d", "1y", "never"
	Notes       string   `json:"notes,omitempty"`
}

type CreateAPIKeyResponse struct {
	APIKey      string     `json:"api_key"`
	KeyID       string     `json:"key_id"`
	ServiceID   string     `json:"service_id"`
	ServiceName string     `json:"service_name"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type APIKeyInfoResponse struct {
	KeyID       string     `json:"key_id"`
	ServiceID   string     `json:"service_id"`
	ServiceName string     `json:"service_name"`
	Permissions []string   `json:"permissions"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	Notes       string     `json:"notes,omitempty"`
}

// CreateAPIKey creates a new API key
// POST /api/admin/keys
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ServiceID == "" {
		http.Error(w, "service_id is required", http.StatusBadRequest)
		return
	}

	// Generate secure API key
	apiKey := generateSecureKey("sk_", 32)
	keyHash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash API key")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresIn != "" && req.ExpiresIn != "never" {
		exp := calculateExpiration(req.ExpiresIn)
		expiresAt = &exp
	}

	// Create API key
	key := &storage.APIKey{
		KeyID:       generateID("key_"),
		KeyHash:     string(keyHash),
		ServiceID:   req.ServiceID,
		ServiceName: req.ServiceName,
		Permissions: req.Permissions,
		ExpiresAt:   expiresAt,
		IsActive:    true,
		Notes:       req.Notes,
		CreatedAt:   time.Now(),
	}

	if err := h.storage.CreateAPIKey(r.Context(), key); err != nil {
		h.logger.WithError(err).Error("Failed to create API key")
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	// Return the API key (only shown once!)
	response := CreateAPIKeyResponse{
		APIKey:      apiKey,
		KeyID:       key.KeyID,
		ServiceID:   key.ServiceID,
		ServiceName: key.ServiceName,
		Permissions: key.Permissions,
		ExpiresAt:   key.ExpiresAt,
		CreatedAt:   key.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	h.logger.WithFields(logrus.Fields{
		"key_id":     key.KeyID,
		"service_id": key.ServiceID,
	}).Info("API key created successfully")
}

// ListAPIKeys lists all API keys
// GET /api/admin/keys
func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"

	keys, err := h.storage.ListAPIKeys(r.Context(), activeOnly)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list API keys")
		http.Error(w, "Failed to list API keys", http.StatusInternalServerError)
		return
	}

	// Convert to response format (without key hashes)
	var response []APIKeyInfoResponse
	for _, key := range keys {
		response = append(response, APIKeyInfoResponse{
			KeyID:       key.KeyID,
			ServiceID:   key.ServiceID,
			ServiceName: key.ServiceName,
			Permissions: key.Permissions,
			CreatedAt:   key.CreatedAt,
			ExpiresAt:   key.ExpiresAt,
			LastUsedAt:  key.LastUsedAt,
			IsActive:    key.IsActive,
			Notes:       key.Notes,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAPIKey retrieves a specific API key
// GET /api/admin/keys/{keyId}
func (h *APIKeyHandler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["keyId"]

	key, err := h.storage.GetAPIKey(r.Context(), keyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get API key")
		http.Error(w, "Failed to get API key", http.StatusInternalServerError)
		return
	}

	if key == nil {
		http.Error(w, "API key not found", http.StatusNotFound)
		return
	}

	response := APIKeyInfoResponse{
		KeyID:       key.KeyID,
		ServiceID:   key.ServiceID,
		ServiceName: key.ServiceName,
		Permissions: key.Permissions,
		CreatedAt:   key.CreatedAt,
		ExpiresAt:   key.ExpiresAt,
		LastUsedAt:  key.LastUsedAt,
		IsActive:    key.IsActive,
		Notes:       key.Notes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RevokeAPIKey revokes an API key
// DELETE /api/admin/keys/{keyId}
func (h *APIKeyHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["keyId"]

	if err := h.storage.RevokeAPIKey(r.Context(), keyID); err != nil {
		h.logger.WithError(err).Error("Failed to revoke API key")
		http.Error(w, "Failed to revoke API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	h.logger.WithField("key_id", keyID).Info("API key revoked")
}

// Helper functions

func generateSecureKey(prefix string, length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	encoded := base64.URLEncoding.EncodeToString(bytes)
	if len(encoded) > length {
		encoded = encoded[:length]
	}
	return prefix + encoded
}

func generateID(prefix string) string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return prefix + base64.URLEncoding.EncodeToString(bytes)[:22]
}

func calculateExpiration(expiresIn string) time.Time {
	now := time.Now()

	switch expiresIn {
	case "30d":
		return now.AddDate(0, 0, 30)
	case "90d":
		return now.AddDate(0, 0, 90)
	case "1y":
		return now.AddDate(1, 0, 0)
	case "2y":
		return now.AddDate(2, 0, 0)
	default:
		return now.AddDate(1, 0, 0) // Default 1 year
	}
}
