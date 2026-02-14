package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Command types
const (
	CmdTypeRefreshAggregates = "refresh_aggregates"
	CmdTypeMetricsStats      = "metrics_stats"
	CmdTypeCleanupOldMetrics = "cleanup_old_metrics"
)

// Command structures
type CommandRequest struct {
	ServerID string                 `json:"server_id"`
	Type     string                 `json:"type"`
	Payload  map[string]interface{} `json:"payload"`
}

type CommandResponse struct {
	CommandID string         `json:"command_id"`
	Status    string         `json:"status"`
	Message   string         `json:"message"`
	Result    *CommandResult `json:"result,omitempty"`
}

type CommandResult struct {
	Success bool                   `json:"success"`
	Output  string                 `json:"output"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Time    time.Time              `json:"time"`
}

// Mock command handlers
func handleRefreshAggregates(req *CommandRequest) *CommandResult {
	granularity, _ := req.Payload["granularity"].(string)
	if granularity == "" {
		granularity = "all"
	}

	return &CommandResult{
		Success: true,
		Output:  fmt.Sprintf("Refreshed %s aggregate successfully", granularity),
		Data: map[string]interface{}{
			"granularity":  granularity,
			"refreshed_at": time.Now(),
		},
		Time: time.Now(),
	}
}

func handleMetricsStats(req *CommandRequest) *CommandResult {
	return &CommandResult{
		Success: true,
		Output:  "Retrieved metrics statistics",
		Data: map[string]interface{}{
			"total_records":  467000,
			"unique_servers": 25,
			"storage_size":   "544 MB",
		},
		Time: time.Now(),
	}
}

func handleCleanupOldMetrics(req *CommandRequest) *CommandResult {
	olderThan, _ := req.Payload["older_than"].(string)
	if olderThan == "" {
		olderThan = "90 days"
	}

	dryRun, _ := req.Payload["dry_run"].(bool)

	if dryRun {
		return &CommandResult{
			Success: true,
			Output:  fmt.Sprintf("DRY RUN: Would delete metrics older than %s", olderThan),
			Data: map[string]interface{}{
				"older_than":     olderThan,
				"dry_run":        true,
				"estimated_rows": 125000,
			},
			Time: time.Now(),
		}
	}

	return &CommandResult{
		Success: true,
		Output:  fmt.Sprintf("Deleted %d metrics older than %s", 125000, olderThan),
		Data: map[string]interface{}{
			"older_than":   olderThan,
			"deleted_rows": 125000,
			"deleted_at":   time.Now(),
		},
		Time: time.Now(),
	}
}

// Main command handler
func handleCommand(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ServerID == "" {
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "command type is required", http.StatusBadRequest)
		return
	}

	// Generate command ID
	commandID := fmt.Sprintf("cmd_%d", time.Now().UnixNano())

	// Execute command based on type
	var result *CommandResult
	switch req.Type {
	case CmdTypeRefreshAggregates:
		result = handleRefreshAggregates(&req)
	case CmdTypeMetricsStats:
		result = handleMetricsStats(&req)
	case CmdTypeCleanupOldMetrics:
		result = handleCleanupOldMetrics(&req)
	default:
		result = &CommandResult{
			Success: false,
			Output:  "",
			Error:   fmt.Sprintf("Unknown command type: %s", req.Type),
			Time:    time.Now(),
		}
	}

	// Determine status
	status := "completed"
	if !result.Success {
		status = "failed"
	}

	// Send response
	response := CommandResponse{
		CommandID: commandID,
		Status:    status,
		Message:   result.Output,
		Result:    result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter()

	// Command endpoint
	router.HandleFunc("/api/servers/{server_id}/command", handleCommand).Methods("POST")

	// Health endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	fmt.Println("🚀 Starting ServerEye API Commands Test Server on :8084")
	fmt.Println("📋 Available endpoints:")
	fmt.Println("  POST /api/servers/{server_id}/command - Execute metrics management command")
	fmt.Println("  GET /health - Health check")
	fmt.Println()
	fmt.Println("🔧 Test commands:")
	fmt.Println(`  curl -X POST http://localhost:8084/api/servers/management/command \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"type":"metrics_stats","payload":{}}'`)

	if err := http.ListenAndServe(":8084", router); err != nil {
		panic(err)
	}
}
