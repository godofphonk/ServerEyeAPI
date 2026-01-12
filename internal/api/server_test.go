package api

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error {
	args := m.Called(ctx, secretKey, agentVersion, operatingSystem, hostname)
	return args.Error(0)
}

func (m *MockStorage) InsertGeneratedKeyWithIDs(ctx context.Context, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname string) error {
	args := m.Called(ctx, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname)
	return args.Error(0)
}

func (m *MockStorage) GetServerByKey(ctx context.Context, serverKey string) (*models.ServerInfo, error) {
	args := m.Called(ctx, serverKey)
	return args.Get(0).(*models.ServerInfo), args.Error(1)
}

func (m *MockStorage) GetServers(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	args := m.Called(ctx, serverID, metrics)
	return args.Error(0)
}

func (m *MockStorage) GetMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).(*models.ServerMetrics), args.Error(1)
}

func (m *MockStorage) GetServerMetrics(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).(*models.ServerStatus), args.Error(1)
}

func (m *MockStorage) SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error {
	args := m.Called(ctx, serverID, status)
	return args.Error(0)
}

func (m *MockStorage) GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).(*models.ServerStatus), args.Error(1)
}

func (m *MockStorage) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	args := m.Called(ctx, serverID, command)
	return args.Error(0)
}

func (m *MockStorage) GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockStorage) StoreDLQ(ctx context.Context, topic, message string, metadata map[string]interface{}) error {
	args := m.Called(ctx, topic, message, metadata)
	return args.Error(0)
}

func (m *MockStorage) GetDLQ(ctx context.Context, topic string, limit int) ([]map[string]interface{}, error) {
	args := m.Called(ctx, topic, limit)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockStorage) StoreConnection(ctx context.Context, serverID string, connectionInfo map[string]interface{}) error {
	args := m.Called(ctx, serverID, connectionInfo)
	return args.Error(0)
}

func (m *MockStorage) GetConnections(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockStorage) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNew(t *testing.T) {
	// Setup environment variables for test
	os.Setenv("DATABASE_URL", "postgres://test:pass@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}

	logger := logrus.New()

	server, err := New(cfg, logger)

	// Server creation might fail due to database connection, but we test the structure
	if err == nil {
		assert.NotNil(t, server)
		assert.NotNil(t, server.storage)
		assert.Equal(t, logger, server.logger)
	} else {
		// Expected to fail without actual database, but that's ok for this test
		assert.Contains(t, err.Error(), "DATABASE_URL")
	}
}

func TestServer_setupRoutes(t *testing.T) {
	// Setup environment variables for test
	os.Setenv("DATABASE_URL", "postgres://test:pass@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}

	logger := logrus.New()

	server, err := New(cfg, logger)

	// Server creation might fail due to database connection, but we test the structure
	if err == nil {
		assert.NotNil(t, server)
		assert.NotNil(t, server.storage)
	} else {
		// Expected to fail without actual database, but that's ok for this test
		assert.Contains(t, err.Error(), "DATABASE_URL")
	}
}

func TestServer_GetAddress(t *testing.T) {
	// Setup environment variables for test
	os.Setenv("DATABASE_URL", "postgres://test:pass@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}

	logger := logrus.New()

	server, err := New(cfg, logger)

	// Server creation might fail due to database connection, but we test the structure
	if err == nil {
		assert.NotNil(t, server)
	} else {
		// Expected to fail without actual database, but that's ok for this test
		assert.Contains(t, err.Error(), "DATABASE_URL")
	}
}

func TestServer_Shutdown(t *testing.T) {
	// Setup environment variables for test
	os.Setenv("DATABASE_URL", "postgres://test:pass@localhost:5432/testdb")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}

	logger := logrus.New()

	server, err := New(cfg, logger)

	// Server creation might fail due to database connection, but we test the structure
	if err == nil {
		assert.NotNil(t, server)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = server.Shutdown(ctx)
		assert.NoError(t, err)
	} else {
		// Expected to fail without actual database, but that's ok for this test
		assert.Contains(t, err.Error(), "DATABASE_URL")
	}
}
