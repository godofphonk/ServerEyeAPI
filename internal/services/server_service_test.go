package services

import (
	"testing"

	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories for testing
type MockServerRepo struct {
	mock.Mock
}

func (m *MockServerRepo) Create(ctx interface{}, server *interfaces.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepo) GetByID(ctx interface{}, id string) (*interfaces.Server, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*interfaces.Server), args.Error(1)
}

func (m *MockServerRepo) GetByKey(ctx interface{}, serverKey string) (*interfaces.Server, error) {
	args := m.Called(ctx, serverKey)
	return args.Get(0).(*interfaces.Server), args.Error(1)
}

func (m *MockServerRepo) Update(ctx interface{}, server *interfaces.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepo) Delete(ctx interface{}, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServerRepo) List(ctx interface{}, opts ...interface{}) ([]*interfaces.Server, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*interfaces.Server), args.Error(1)
}

func (m *MockServerRepo) Ping(ctx interface{}) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockKeyRepo struct {
	mock.Mock
}

func (m *MockKeyRepo) Create(ctx interface{}, key *interfaces.GeneratedKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockKeyRepo) GetByKey(ctx interface{}, serverKey string) (*interfaces.GeneratedKey, error) {
	args := m.Called(ctx, serverKey)
	return args.Get(0).(*interfaces.GeneratedKey), args.Error(1)
}

func (m *MockKeyRepo) GetByServerID(ctx interface{}, serverID string) (*interfaces.GeneratedKey, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).(*interfaces.GeneratedKey), args.Error(1)
}

func (m *MockKeyRepo) Update(ctx interface{}, key *interfaces.GeneratedKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockKeyRepo) Delete(ctx interface{}, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockKeyRepo) List(ctx interface{}, opts ...interface{}) ([]*interfaces.GeneratedKey, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*interfaces.GeneratedKey), args.Error(1)
}

func (m *MockKeyRepo) Ping(ctx interface{}) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewServerService(t *testing.T) {
	mockServerRepo := &MockServerRepo{}
	mockKeyRepo := &MockKeyRepo{}
	logger := logrus.New()

	service := NewServerService(mockServerRepo, mockKeyRepo, logger)

	assert.NotNil(t, service)
}
