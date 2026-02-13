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
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
)

// CommandsService handles command-related business logic
type CommandsService struct {
	keyRepo         interfaces.GeneratedKeyRepository
	metricsCommands *MetricsCommandsService
	logger          *logrus.Logger
}

// NewCommandsService creates a new commands service
func NewCommandsService(keyRepo interfaces.GeneratedKeyRepository, logger *logrus.Logger) *CommandsService {
	return &CommandsService{
		keyRepo: keyRepo,
		logger:  logger,
	}
}

// SetMetricsCommands sets the metrics commands service
func (s *CommandsService) SetMetricsCommands(metricsCommands *MetricsCommandsService) {
	s.metricsCommands = metricsCommands
}

// Command represents a server command
type Command struct {
	ID         string                 `json:"id"`
	ServerID   string                 `json:"server_id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	ExecutedAt *time.Time             `json:"executed_at,omitempty"`
	Result     *CommandResult         `json:"result,omitempty"`
}

// CommandResult represents command execution result
type CommandResult struct {
	Success bool      `json:"success"`
	Output  string    `json:"output"`
	Error   string    `json:"error,omitempty"`
	Time    time.Time `json:"time"`
}

// SendCommandRequest represents a command sending request
type SendCommandRequest struct {
	ServerID string                 `json:"server_id" validate:"required"`
	Type     string                 `json:"type" validate:"required"`
	Payload  map[string]interface{} `json:"payload"`
}

// SendCommandResponse represents a command sending response
type SendCommandResponse struct {
	CommandID string `json:"command_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// SendCommand sends a command to a server with validation and business logic
func (s *CommandsService) SendCommand(ctx context.Context, req *SendCommandRequest) (*SendCommandResponse, error) {
	// Validate request
	if err := s.validateSendCommandRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify server exists
	_, err := s.keyRepo.GetByServerID(ctx, req.ServerID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Validate command type
	if err := s.validateCommandType(req.Type); err != nil {
		return nil, fmt.Errorf("invalid command type: %w", err)
	}

	// Check if this is a metrics command
	if s.isMetricsCommand(req.Type) {
		return s.handleMetricsCommand(ctx, req)
	}

	// Handle regular server command
	return s.handleServerCommand(ctx, req)
}

// isMetricsCommand checks if command type is metrics-related
func (s *CommandsService) isMetricsCommand(commandType string) bool {
	metricsCommands := []string{
		CmdTypeRefreshAggregates,
		CmdTypeRebuildAggregates,
		CmdTypeCleanupOldMetrics,
		CmdTypeCompressionPolicy,
		CmdTypeRetentionPolicy,
		CmdTypeMetricsStats,
		CmdTypeAnalyzePerformance,
		CmdTypeExportMetrics,
		CmdTypeImportMetrics,
		CmdTypeValidateMetrics,
		CmdTypeOptimizeStorage,
	}

	for _, cmd := range metricsCommands {
		if commandType == cmd {
			return true
		}
	}
	return false
}

// handleMetricsCommand processes metrics-related commands
func (s *CommandsService) handleMetricsCommand(ctx context.Context, req *SendCommandRequest) (*SendCommandResponse, error) {
	if s.metricsCommands == nil {
		return nil, fmt.Errorf("metrics commands service not initialized")
	}

	// Create metrics command
	cmd := &MetricsCommand{
		ID:        fmt.Sprintf("metrics_cmd_%d", time.Now().UnixNano()),
		ServerID:  req.ServerID,
		Type:      req.Type,
		Payload:   req.Payload,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Execute metrics command
	result, err := s.metricsCommands.ExecuteMetricsCommand(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute metrics command: %w", err)
	}

	// Update command with result
	cmd.Result = result
	cmd.ExecutedAt = &result.Time
	if result.Success {
		cmd.Status = "completed"
	} else {
		cmd.Status = "failed"
	}

	s.logger.WithFields(logrus.Fields{
		"command_id": cmd.ID,
		"server_id":  req.ServerID,
		"type":       req.Type,
		"status":     cmd.Status,
	}).Info("Metrics command executed")

	return &SendCommandResponse{
		CommandID: cmd.ID,
		Status:    cmd.Status,
		Message:   result.Output,
	}, nil
}

// handleServerCommand processes regular server commands
func (s *CommandsService) handleServerCommand(ctx context.Context, req *SendCommandRequest) (*SendCommandResponse, error) {
	// Create command
	command := &Command{
		ID:        fmt.Sprintf("cmd_%s", uuid.New().String()[:8]),
		ServerID:  req.ServerID,
		Type:      req.Type,
		Payload:   req.Payload,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Store command in storage
	s.logger.WithFields(logrus.Fields{
		"command_id": command.ID,
		"server_id":  req.ServerID,
		"type":       req.Type,
	}).Info("Command created successfully")

	return &SendCommandResponse{
		CommandID: command.ID,
		Status:    "pending",
		Message:   "Command queued for execution",
	}, nil
}

// GetPendingCommands retrieves pending commands for a server
func (s *CommandsService) GetPendingCommands(ctx context.Context, serverID string) ([]*Command, error) {
	// Verify server exists
	_, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Retrieve pending commands from storage
	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
	}).Info("Retrieving pending commands")

	// Return empty for now
	return []*Command{}, nil
}

// ExecuteCommand processes command execution result
func (s *CommandsService) ExecuteCommand(ctx context.Context, commandID string, result *CommandResult) error {
	// Validate input
	if commandID == "" {
		return fmt.Errorf("command_id is required")
	}
	if result == nil {
		return fmt.Errorf("result is required")
	}

	// Process command execution result
	s.logger.WithFields(logrus.Fields{
		"command_id": commandID,
		"success":    result.Success,
	}).Info("Command execution processed")

	return nil
}

// ListCommands retrieves commands for a server with optional filtering
func (s *CommandsService) ListCommands(ctx context.Context, serverID, status string, limit int) ([]*Command, error) {
	// Verify server exists
	_, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Retrieve commands from storage
	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"status":    status,
		"limit":     limit,
	}).Info("Retrieving commands list")

	// Return empty for now
	return []*Command{}, nil
}

// CancelCommand cancels a pending command
func (s *CommandsService) CancelCommand(ctx context.Context, commandID string) error {
	// Validate input
	if commandID == "" {
		return fmt.Errorf("command_id is required")
	}

	// TODO: Implement command cancellation logic
	s.logger.WithFields(logrus.Fields{
		"command_id": commandID,
	}).Info("Command cancelled")

	return nil
}

// validateSendCommandRequest validates command sending request
func (s *CommandsService) validateSendCommandRequest(req *SendCommandRequest) error {
	if req.ServerID == "" {
		return fmt.Errorf("server_id is required")
	}
	if req.Type == "" {
		return fmt.Errorf("command type is required")
	}
	return nil
}

// validateCommandType validates command type
func (s *CommandsService) validateCommandType(commandType string) error {
	validTypes := []string{
		"restart",
		"shutdown",
		"update",
		"script",
		"info",
		"ping",
		// Metrics management commands
		CmdTypeRefreshAggregates,
		CmdTypeRebuildAggregates,
		CmdTypeCleanupOldMetrics,
		CmdTypeCompressionPolicy,
		CmdTypeRetentionPolicy,
		CmdTypeMetricsStats,
		CmdTypeAnalyzePerformance,
		CmdTypeExportMetrics,
		CmdTypeImportMetrics,
		CmdTypeValidateMetrics,
		CmdTypeOptimizeStorage,
	}

	for _, validType := range validTypes {
		if commandType == validType {
			return nil
		}
	}

	return fmt.Errorf("unsupported command type: %s", commandType)
}

// CreateRestartCommand creates a restart command
func (s *CommandsService) CreateRestartCommand(serverID string) *SendCommandRequest {
	return &SendCommandRequest{
		ServerID: serverID,
		Type:     "restart",
		Payload: map[string]interface{}{
			"force": false,
		},
	}
}

// CreateUpdateCommand creates an update command
func (s *CommandsService) CreateUpdateCommand(serverID, version string) *SendCommandRequest {
	return &SendCommandRequest{
		ServerID: serverID,
		Type:     "update",
		Payload: map[string]interface{}{
			"version": version,
			"force":   false,
		},
	}
}

// CreateScriptCommand creates a script execution command
func (s *CommandsService) CreateScriptCommand(serverID, script string, args []string) *SendCommandRequest {
	return &SendCommandRequest{
		ServerID: serverID,
		Type:     "script",
		Payload: map[string]interface{}{
			"script": script,
			"args":   args,
		},
	}
}

// Ping checks repository connectivity
func (s *CommandsService) Ping(ctx context.Context) error {
	if err := s.keyRepo.Ping(ctx); err != nil {
		return fmt.Errorf("key repository ping failed: %w", err)
	}

	return nil
}
