package models

// RegisterKeyRequest represents request to register a key
type RegisterKeyRequest struct {
	SecretKey       string `json:"secret_key"`
	AgentVersion    string `json:"agent_version"`
	OperatingSystem string `json:"operating_system"`
	Hostname        string `json:"hostname"`
}

// SendCommandRequest represents request to send command to server
type SendCommandRequest struct {
	ServerID string                 `json:"server_id"`
	Command  map[string]interface{} `json:"command"`
}
