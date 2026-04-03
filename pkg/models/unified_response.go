package models

// UnifiedServerResponse combines metrics, status, and static info in one response
type UnifiedServerResponse struct {
	ServerID  string `json:"server_id"`
	ServerKey string `json:"server_key,omitempty"`
	Timestamp string `json:"timestamp"`

	// Metrics data (from /metrics endpoint - current server state)
	Metrics interface{} `json:"metrics,omitempty"`

	// Status data (from /status endpoint)
	Status interface{} `json:"status,omitempty"`

	// Static info data (from /static-info endpoint)
	StaticInfo interface{} `json:"static_info,omitempty"`

	// Performance metadata
	ResponseMeta ResponseMetadata `json:"response_meta"`
}

// ResponseMetadata provides information about the unified response
type ResponseMetadata struct {
	TotalResponseTimeMs int64                      `json:"total_response_time_ms"`
	ComponentsStatus    map[string]ComponentStatus `json:"components_status"`
	DataPointsCount     map[string]int             `json:"data_points_count"`
}

// ComponentStatus indicates the status of each component
type ComponentStatus struct {
	Available    bool   `json:"available"`
	ResponseTime int64  `json:"response_time_ms"`
	Error        string `json:"error,omitempty"`
}

// ServerMetricsResponse represents the metrics component
type ServerMetricsResponse struct {
	ServerID       string      `json:"server_id"`
	StartTime      string      `json:"start_time"`
	EndTime        string      `json:"end_time"`
	Granularity    string      `json:"granularity"`
	DataPoints     []DataPoint `json:"data_points"`
	TotalPoints    int         `json:"total_points"`
	NetworkDetails NetworkInfo `json:"network_details"`
	DiskDetails    DiskInfo    `json:"disk_details"`
	TempDetails    TempInfo    `json:"temperature_details"`
}

// ServerStatusResponse represents the status component
type ServerStatusResponse struct {
	ServerID      string  `json:"server_id"`
	Status        string  `json:"status"`
	LastSeen      string  `json:"last_seen"`
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	DiskUsage     float64 `json:"disk_usage"`
	NetworkStatus string  `json:"network_status"`
	IsOnline      bool    `json:"is_online"`
	Alerts        []Alert `json:"alerts,omitempty"`
}

// StaticInfoResponse represents the static info component
type StaticInfoResponse struct {
	ServerInfo *ServerStaticInfo  `json:"server_info"`
	Hardware   *HardwareInfo      `json:"hardware"`
	Network    *NetworkStaticInfo `json:"network"`
	Storage    *StorageInfo       `json:"storage"`
	System     *SystemInfo        `json:"system"`
}

// Reusing existing models from other files...
type DataPoint struct {
	Timestamp   string  `json:"timestamp"`
	CPUAvg      float64 `json:"cpu_avg"`
	CPUMax      float64 `json:"cpu_max"`
	CPUMin      float64 `json:"cpu_min"`
	MemoryAvg   float64 `json:"memory_avg"`
	MemoryMax   float64 `json:"memory_max"`
	MemoryMin   float64 `json:"memory_min"`
	DiskAvg     float64 `json:"disk_avg"`
	DiskMax     float64 `json:"disk_max"`
	NetworkAvg  float64 `json:"network_avg"`
	NetworkMax  float64 `json:"network_max"`
	TempAvg     float64 `json:"temp_avg"`
	TempMax     float64 `json:"temp_max"`
	LoadAvg     float64 `json:"load_avg"`
	LoadMax     float64 `json:"load_max"`
	SampleCount int     `json:"sample_count"`
}

type NetworkInfo struct {
	Interfaces  []NetworkInterfaceInfo `json:"interfaces"`
	TotalRxMbps float64                `json:"total_rx_mbps"`
	TotalTxMbps float64                `json:"total_tx_mbps"`
}

type DiskInfo struct {
	Disks []DiskDriveInfo `json:"disks"`
}

type TempInfo struct {
	CPUTemperature      float64            `json:"cpu_temperature"`
	GPUTemperature      float64            `json:"gpu_temperature"`
	SystemTemperature   float64            `json:"system_temperature"`
	StorageTemperatures map[string]float64 `json:"storage_temperatures"`
	HighestTemperature  float64            `json:"highest_temperature"`
	TemperatureUnit     string             `json:"temperature_unit"`
}

type Alert struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
	ResolvedAt string `json:"resolved_at,omitempty"`
}

// Static info types (simplified versions)
type ServerStaticInfo struct {
	ServerID      string `json:"server_id"`
	Hostname      string `json:"hostname"`
	OSInfo        string `json:"os_info"`
	Kernel        string `json:"kernel"`
	Architecture  string `json:"architecture"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	UptimeHuman   string `json:"uptime_human"`
	BootTime      string `json:"boot_time"`
	AgentVersion  string `json:"agent_version"`
}

type HardwareInfo struct {
	CPU         CPUInfo         `json:"cpu"`
	Memory      MemoryInfo      `json:"memory"`
	Motherboard MotherboardInfo `json:"motherboard"`
}

type NetworkStaticInfo struct {
	Interfaces []NetworkInterfaceStatic `json:"interfaces"`
	Routes     []Route                  `json:"routes"`
	DNS        []DNS                    `json:"dns"`
}

type StorageInfo struct {
	Disks []DiskStatic `json:"disks"`
	Raids []RAID       `json:"raids,omitempty"`
}

type SystemInfo struct {
	OS        OSInfo      `json:"os"`
	Processes ProcessInfo `json:"processes"`
	Services  []Service   `json:"services"`
}

// Placeholder types for compilation
type CPUInfo struct {
	Model     string  `json:"model"`
	Cores     int     `json:"cores"`
	Threads   int     `json:"threads"`
	Frequency float64 `json:"frequency"`
}

type MemoryInfo struct {
	TotalGB     float64 `json:"total_gb"`
	AvailableGB float64 `json:"available_gb"`
	Type        string  `json:"type"`
}

type MotherboardInfo struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Version      string `json:"version"`
}

type NetworkInterfaceStatic struct {
	Name   string   `json:"name"`
	Status string   `json:"status"`
	MAC    string   `json:"mac"`
	IPs    []string `json:"ips"`
	Speed  int      `json:"speed"`
	Duplex string   `json:"duplex"`
}

type Route struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
}

type DNS struct {
	Server string `json:"server"`
	Type   string `json:"type"`
}

type DiskStatic struct {
	Path   string  `json:"path"`
	Type   string  `json:"type"`
	SizeGB float64 `json:"size_gb"`
	Model  string  `json:"model"`
	Serial string  `json:"serial"`
}

type RAID struct {
	Name   string `json:"name"`
	Level  string `json:"level"`
	Status string `json:"status"`
}

type OSInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Build    string `json:"build"`
	Platform string `json:"platform"`
}

type ProcessInfo struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Sleeping int `json:"sleeping"`
}

type Service struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	State  string `json:"state"`
}

// Additional types needed for compilation
type NetworkInterfaceInfo struct {
	Name       string `json:"name"`
	MAC        string `json:"mac"`
	IP         string `json:"ip"`
	Status     string `json:"status"`
	SpeedMbps  int    `json:"speed_mbps"`
	Type       string `json:"type"`
	IsPhysical bool   `json:"is_physical"`
}

type DiskDriveInfo struct {
	Name         string  `json:"name"`
	Model        string  `json:"model"`
	SerialNumber string  `json:"serial_number"`
	SizeGB       float64 `json:"size_gb"`
	Type         string  `json:"type"`
	Interface    string  `json:"interface"`
	MountPoint   string  `json:"mount_point"`
	IsSystemDisk bool    `json:"is_system_disk"`
}
