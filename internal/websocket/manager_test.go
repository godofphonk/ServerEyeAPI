package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.clients)
	assert.NotNil(t, manager.broadcast)
	assert.NotNil(t, manager.register)
	assert.NotNil(t, manager.unregister)
}

func TestManager_GetClientCount(t *testing.T) {
	manager := NewManager()

	// Initially should have 0 clients
	count := manager.GetClientCount()
	assert.Equal(t, 0, count)
}

func TestManager_Shutdown(t *testing.T) {
	manager := NewManager()

	// Shutdown should not panic
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestManager_GetClients(t *testing.T) {
	manager := NewManager()

	// Initially should have no clients
	clients := manager.GetClients()
	assert.Equal(t, 0, len(clients))
}
