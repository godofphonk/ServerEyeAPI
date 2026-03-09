package handlers

import "github.com/godofphonk/ServerEyeAPI/internal/models"

// convertV2ToOldFormat converts new MetricsV2 format to old ServerMetrics format
func (h *MetricsPushHandler) convertV2ToOldFormat(v2 *models.MetricsV2) *models.ServerMetrics {
	old := &models.ServerMetrics{}

	// Aggregated values for backward compatibility
	old.CPU = v2.CPUUsage.UsageTotal
	old.Memory = v2.Memory.UsedPercent
	
	// Calculate average disk usage
	if len(v2.Disks) > 0 {
		var totalDiskUsage float64
		for _, disk := range v2.Disks {
			totalDiskUsage += disk.UsedPercent
		}
		old.Disk = totalDiskUsage / float64(len(v2.Disks))
	}
	
	// Calculate total network traffic in MB
	var totalRxMB, totalTxMB float64
	for _, iface := range v2.Network.Interfaces {
		totalRxMB += float64(iface.RxBytes) / 1024 / 1024
		totalTxMB += float64(iface.TxBytes) / 1024 / 1024
	}
	old.Network = totalRxMB + totalTxMB

	// CPU detailed metrics
	old.CPUUsage.UsageTotal = v2.CPUUsage.UsageTotal
	old.CPUUsage.UsageUser = v2.CPUUsage.UsageUser
	old.CPUUsage.UsageSystem = v2.CPUUsage.UsageSystem
	old.CPUUsage.UsageIdle = v2.CPUUsage.UsageIdle
	old.CPUUsage.LoadAverage.Load1 = v2.CPUUsage.LoadAverage.Load1Min
	old.CPUUsage.LoadAverage.Load5 = v2.CPUUsage.LoadAverage.Load5Min
	old.CPUUsage.LoadAverage.Load15 = v2.CPUUsage.LoadAverage.Load15Min
	old.CPUUsage.Frequency = v2.CPUUsage.FrequencyMHz

	// Memory detailed metrics
	old.MemoryDetails.TotalGB = v2.Memory.TotalGB
	old.MemoryDetails.UsedGB = v2.Memory.UsedGB
	old.MemoryDetails.AvailableGB = v2.Memory.AvailableGB
	old.MemoryDetails.FreeGB = v2.Memory.FreeGB
	old.MemoryDetails.BuffersGB = v2.Memory.BuffersGB
	old.MemoryDetails.CachedGB = v2.Memory.CachedGB
	old.MemoryDetails.UsedPercent = v2.Memory.UsedPercent

	// Disk detailed metrics
	if len(v2.Disks) > 0 {
		old.DiskDetails = make([]struct {
			Path        string  `json:"path"`
			TotalGB     float64 `json:"total_gb"`
			UsedGB      float64 `json:"used_gb"`
			FreeGB      float64 `json:"free_gb"`
			UsedPercent float64 `json:"used_percent"`
			Filesystem  string  `json:"filesystem"`
		}, len(v2.Disks))
		for i, disk := range v2.Disks {
			old.DiskDetails[i].Path = disk.MountPoint
			old.DiskDetails[i].UsedGB = disk.UsedGB
			old.DiskDetails[i].FreeGB = disk.FreeGB
			old.DiskDetails[i].UsedPercent = disk.UsedPercent
			old.DiskDetails[i].TotalGB = disk.UsedGB + disk.FreeGB
		}
	}

	// Network detailed metrics
	if len(v2.Network.Interfaces) > 0 {
		old.NetworkDetails.Interfaces = make([]struct {
			Name        string  `json:"name"`
			RxBytes     int64   `json:"rx_bytes"`
			TxBytes     int64   `json:"tx_bytes"`
			RxPackets   int64   `json:"rx_packets"`
			TxPackets   int64   `json:"tx_packets"`
			RxSpeedMbps float64 `json:"rx_speed_mbps"`
			TxSpeedMbps float64 `json:"tx_speed_mbps"`
			Status      string  `json:"status"`
		}, len(v2.Network.Interfaces))
		for i, iface := range v2.Network.Interfaces {
			old.NetworkDetails.Interfaces[i].Name = iface.Name
			old.NetworkDetails.Interfaces[i].RxBytes = iface.RxBytes
			old.NetworkDetails.Interfaces[i].TxBytes = iface.TxBytes
			old.NetworkDetails.Interfaces[i].RxPackets = iface.RxPackets
			old.NetworkDetails.Interfaces[i].TxPackets = iface.TxPackets
			old.NetworkDetails.Interfaces[i].RxSpeedMbps = iface.RxSpeedMbps
			old.NetworkDetails.Interfaces[i].TxSpeedMbps = iface.TxSpeedMbps
			old.NetworkDetails.Interfaces[i].Status = iface.Status
		}
		old.NetworkDetails.TotalRxMbps = v2.Network.TotalRxMbps
		old.NetworkDetails.TotalTxMbps = v2.Network.TotalTxMbps
	}

	// Temperature detailed metrics
	old.TemperatureDetails.CPUTemperature = v2.Temperature.CPU
	old.TemperatureDetails.GPUTemperature = v2.Temperature.GPU
	old.TemperatureDetails.HighestTemperature = v2.Temperature.Highest

	if len(v2.Temperature.Storage) > 0 {
		old.TemperatureDetails.StorageTemperatures = make([]struct {
			Device      string  `json:"device"`
			Type        string  `json:"type"`
			Temperature float64 `json:"temperature"`
		}, len(v2.Temperature.Storage))

		for i, storage := range v2.Temperature.Storage {
			old.TemperatureDetails.StorageTemperatures[i].Device = storage.Device
			old.TemperatureDetails.StorageTemperatures[i].Temperature = storage.Temperature
		}
	}

	// System detailed metrics
	old.SystemDetails.ProcessesTotal = v2.System.ProcessesTotal
	old.SystemDetails.ProcessesRunning = v2.System.ProcessesRunning
	old.SystemDetails.ProcessesSleeping = v2.System.ProcessesSleeping
	old.SystemDetails.UptimeSeconds = v2.System.UptimeSeconds

	return old
}
