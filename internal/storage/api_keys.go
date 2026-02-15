package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type APIKey struct {
	KeyID       string
	KeyHash     string
	ServiceID   string
	ServiceName string
	Permissions []string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	LastUsedAt  *time.Time
	IsActive    bool
	CreatedBy   string
	Notes       string
}

type APIKeyStorage struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewAPIKeyStorage(db *sql.DB, logger *logrus.Logger) *APIKeyStorage {
	return &APIKeyStorage{
		db:     db,
		logger: logger,
	}
}

// CreateAPIKey creates a new API key in the database
func (s *APIKeyStorage) CreateAPIKey(ctx context.Context, key *APIKey) error {
	query := `
		INSERT INTO api_keys (
			key_id, key_hash, service_id, service_name, 
			permissions, expires_at, is_active, created_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.db.ExecContext(ctx, query,
		key.KeyID,
		key.KeyHash,
		key.ServiceID,
		key.ServiceName,
		pq.Array(key.Permissions),
		key.ExpiresAt,
		key.IsActive,
		key.CreatedBy,
		key.Notes,
	)

	if err != nil {
		s.logger.WithError(err).Error("Failed to create API key")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"key_id":     key.KeyID,
		"service_id": key.ServiceID,
	}).Info("API key created successfully")

	return nil
}

// ValidateAPIKey validates an API key and returns the key details if valid
func (s *APIKeyStorage) ValidateAPIKey(ctx context.Context, apiKey string) (*APIKey, error) {
	query := `
		SELECT key_id, key_hash, service_id, service_name, permissions, 
		       created_at, expires_at, last_used_at, is_active, created_by, notes
		FROM api_keys
		WHERE is_active = true
		  AND (expires_at IS NULL OR expires_at > NOW())
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Failed to query API keys")
		return nil, err
	}
	defer rows.Close()

	// Check each active key
	for rows.Next() {
		var key APIKey
		var permissions pq.StringArray

		err := rows.Scan(
			&key.KeyID,
			&key.KeyHash,
			&key.ServiceID,
			&key.ServiceName,
			&permissions,
			&key.CreatedAt,
			&key.ExpiresAt,
			&key.LastUsedAt,
			&key.IsActive,
			&key.CreatedBy,
			&key.Notes,
		)
		if err != nil {
			continue
		}

		key.Permissions = permissions

		// Compare the provided key with the stored hash
		err = bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(apiKey))
		if err == nil {
			// Key is valid
			return &key, nil
		}
	}

	return nil, sql.ErrNoRows
}

// UpdateLastUsed updates the last_used_at timestamp for an API key
func (s *APIKeyStorage) UpdateLastUsed(ctx context.Context, keyID string) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE key_id = $1`

	_, err := s.db.ExecContext(ctx, query, keyID)
	if err != nil {
		s.logger.WithError(err).WithField("key_id", keyID).Error("Failed to update last_used_at")
		return err
	}

	return nil
}

// LogAPIKeyUsage logs API key usage to audit log
func (s *APIKeyStorage) LogAPIKeyUsage(ctx context.Context, keyID, endpoint, ipAddress, userAgent string, success bool, errorMessage string) error {
	query := `
		INSERT INTO api_key_audit_log (key_id, endpoint, ip_address, user_agent, success, error_message)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, keyID, endpoint, ipAddress, userAgent, success, errorMessage)
	if err != nil {
		s.logger.WithError(err).Error("Failed to log API key usage")
		return err
	}

	return nil
}

// RevokeAPIKey deactivates an API key
func (s *APIKeyStorage) RevokeAPIKey(ctx context.Context, keyID string) error {
	query := `UPDATE api_keys SET is_active = false WHERE key_id = $1`

	_, err := s.db.ExecContext(ctx, query, keyID)
	if err != nil {
		s.logger.WithError(err).WithField("key_id", keyID).Error("Failed to revoke API key")
		return err
	}

	s.logger.WithField("key_id", keyID).Info("API key revoked")
	return nil
}

// GetAPIKey retrieves an API key by ID
func (s *APIKeyStorage) GetAPIKey(ctx context.Context, keyID string) (*APIKey, error) {
	query := `
		SELECT key_id, key_hash, service_id, service_name, permissions, 
		       created_at, expires_at, last_used_at, is_active, created_by, notes
		FROM api_keys
		WHERE key_id = $1
	`

	var key APIKey
	var permissions pq.StringArray

	err := s.db.QueryRowContext(ctx, query, keyID).Scan(
		&key.KeyID,
		&key.KeyHash,
		&key.ServiceID,
		&key.ServiceName,
		&permissions,
		&key.CreatedAt,
		&key.ExpiresAt,
		&key.LastUsedAt,
		&key.IsActive,
		&key.CreatedBy,
		&key.Notes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.logger.WithError(err).WithField("key_id", keyID).Error("Failed to get API key")
		return nil, err
	}

	key.Permissions = permissions
	return &key, nil
}

// ListAPIKeys returns all API keys
func (s *APIKeyStorage) ListAPIKeys(ctx context.Context, activeOnly bool) ([]*APIKey, error) {
	query := `
		SELECT key_id, key_hash, service_id, service_name, permissions, 
		       created_at, expires_at, last_used_at, is_active, created_by, notes
		FROM api_keys
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list API keys")
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var key APIKey
		var permissions pq.StringArray

		err := rows.Scan(
			&key.KeyID,
			&key.KeyHash,
			&key.ServiceID,
			&key.ServiceName,
			&permissions,
			&key.CreatedAt,
			&key.ExpiresAt,
			&key.LastUsedAt,
			&key.IsActive,
			&key.CreatedBy,
			&key.Notes,
		)
		if err != nil {
			continue
		}

		key.Permissions = permissions
		keys = append(keys, &key)
	}

	return keys, nil
}
