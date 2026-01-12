package models

// ServerInfo represents server information for authentication
type ServerInfo struct {
	ServerID  string `json:"server_id"`
	SecretKey string `json:"secret_key"`
	Hostname  string `json:"hostname"`
}
