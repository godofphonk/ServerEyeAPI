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

package services

import (
	"context"
	"testing"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories for testing
type MockServerRepo struct {
	mock.Mock
}

func (m *MockServerRepo) Create(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepo) GetByID(ctx context.Context, id string) (*models.Server, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *MockServerRepo) GetByKey(ctx context.Context, serverKey string) (*models.Server, error) {
	args := m.Called(ctx, serverKey)
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *MockServerRepo) Update(ctx context.Context, server *models.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServerRepo) List(ctx context.Context, opts ...interfaces.ListOption) ([]*models.Server, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*models.Server), args.Error(1)
}

func (m *MockServerRepo) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockServerRepo) UpdateSources(ctx context.Context, serverID string, sources string) error {
	args := m.Called(ctx, serverID, sources)
	return args.Error(0)
}

func (m *MockServerRepo) ListByStatus(ctx context.Context, status string) ([]*models.Server, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*models.Server), args.Error(1)
}

func (m *MockServerRepo) ListByHostname(ctx context.Context, hostname string) ([]*models.Server, error) {
	args := m.Called(ctx, hostname)
	return args.Get(0).([]*models.Server), args.Error(1)
}

func (m *MockServerRepo) UpdateStatus(ctx context.Context, serverID string, status string) error {
	args := m.Called(ctx, serverID, status)
	return args.Error(0)
}

func (m *MockServerRepo) UpdateLastSeen(ctx context.Context, serverID string, lastSeen time.Time) error {
	args := m.Called(ctx, serverID, lastSeen)
	return args.Error(0)
}

type MockKeyRepo struct {
	mock.Mock
}

func (m *MockKeyRepo) Create(ctx context.Context, key *models.GeneratedKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockKeyRepo) GetByKey(ctx context.Context, serverKey string) (*models.GeneratedKey, error) {
	args := m.Called(ctx, serverKey)
	return args.Get(0).(*models.GeneratedKey), args.Error(1)
}

func (m *MockKeyRepo) GetByServerID(ctx context.Context, serverID string) (*models.GeneratedKey, error) {
	args := m.Called(ctx, serverID)
	return args.Get(0).(*models.GeneratedKey), args.Error(1)
}

func (m *MockKeyRepo) Update(ctx context.Context, key *models.GeneratedKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockKeyRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockKeyRepo) List(ctx context.Context, opts ...interfaces.ListOption) ([]*models.GeneratedKey, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*models.GeneratedKey), args.Error(1)
}

func (m *MockKeyRepo) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKeyRepo) ListByStatus(ctx context.Context, status string) ([]*models.GeneratedKey, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*models.GeneratedKey), args.Error(1)
}

func TestNewServerService(t *testing.T) {
	mockServerRepo := &MockServerRepo{}
	mockKeyRepo := &MockKeyRepo{}
	logger := logrus.New()

	service := NewServerService(mockServerRepo, mockKeyRepo, logger)

	assert.NotNil(t, service)
}
