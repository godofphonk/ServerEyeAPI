package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
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

func (m *MockStorage) GetServerByKey(ctx context.Context, serverKey string) (*storage.ServerInfo, error) {
	args := m.Called(ctx, serverKey)
	return args.Get(0).(*storage.ServerInfo), args.Error(1)
}

func (m *MockStorage) GetServers(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) StoreMetric(ctx context.Context, serverID string, metrics *storage.ServerMetrics) error {
	args := m.Called(ctx, serverID, metrics)
	return args.Error(0)
}

func (m *MockStorage) GetMetric(ctx context.Context, serverID string) (*storage.ServerMetrics, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).(*storage.ServerMetrics), args.Error(1)
}

func (m *MockStorage) GetServerMetrics(ctx context.Context, serverID string) (*storage.ServerStatus, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).(*storage.ServerStatus), args.Error(1)
}

func (m *MockStorage) SetServerStatus(ctx context.Context, serverID string, status *storage.ServerStatus) error {
	args := m.Called(ctx, serverID, status)
	return args.Error(0)
}

func (m *MockStorage) GetServerStatus(ctx context.Context, serverID string) (*storage.ServerStatus, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).(*storage.ServerStatus), args.Error(1)
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
	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}

	mockStorage := &MockStorage{}
	logger := logrus.New()

	server, err := New(cfg, mockStorage, logger)

	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.Equal(t, mockStorage, server.storage)
	assert.Equal(t, logger, server.logger)
}

func TestServer_setupRoutes(t *testing.T) {
	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}

	mockStorage := &MockStorage{}
	logger := logrus.New()

	server, err := New(cfg, mockStorage, logger)
	assert.NoError(t, err)

	router := server.setupRoutes()
	assert.NotNil(t, router)

	// Test health route
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestServer_GetAddress(t *testing.T) {
	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}

	mockStorage := &MockStorage{}
	logger := logrus.New()

	server, err := New(cfg, mockStorage, logger)
	assert.NoError(t, err)

	address := server.GetAddress()
	assert.Equal(t, "localhost:8080", address)
}

func TestServer_Shutdown(t *testing.T) {
	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}

	mockStorage := &MockStorage{}
	mockStorage.On("Close").Return(nil)

	logger := logrus.New()

	server, err := New(cfg, mockStorage, logger)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}
