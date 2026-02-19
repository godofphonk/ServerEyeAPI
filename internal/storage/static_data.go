package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// StaticDataStorage handles static/persistent server information
type StaticDataStorage interface {
	// Server Info
	UpsertServerInfo(ctx context.Context, info *ServerInfo) error
	GetServerInfo(ctx context.Context, serverID string) (*ServerInfo, error)
	
	// Hardware Info
	UpsertHardwareInfo(ctx context.Context, info *HardwareInfo) error
	GetHardwareInfo(ctx context.Context, serverID string) (*HardwareInfo, error)
	
	// Network Interfaces
	UpsertNetworkInterfaces(ctx context.Context, serverID string, interfaces []NetworkInterface) error
	GetNetworkInterfaces(ctx context.Context, serverID string) ([]NetworkInterface, error)
	
	// Disk Info
	UpsertDiskInfo(ctx context.Context, serverID string, disks []DiskInfo) error
	GetDiskInfo(ctx context.Context, serverID string) ([]DiskInfo, error)
	
	// Combined operations
	GetCompleteStaticInfo(ctx context.Context, serverID string) (*CompleteStaticInfo, error)
	UpsertCompleteStaticInfo(ctx context.Context, serverID string, info *CompleteStaticInfo) error
}

// ServerInfo represents basic server system information
type ServerInfo struct {
	ServerID     string    `json:"server_id"`
	Hostname     string    `json:"hostname"`
	OS           string    `json:"os"`
	OSVersion    string    `json:"os_version"`
	Kernel       string    `json:"kernel"`
	Architecture string    `json:"architecture"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// HardwareInfo represents hardware specifications
type HardwareInfo struct {
	ServerID        string    `json:"server_id"`
	CPUModel        string    `json:"cpu_model"`
	CPUCores        int       `json:"cpu_cores"`
	CPUThreads      int       `json:"cpu_threads"`
	CPUFrequencyMHz int       `json:"cpu_frequency_mhz"`
	GPUModel        string    `json:"gpu_model"`
	GPUDriver       string    `json:"gpu_driver"`
	GPUMemoryGB     int       `json:"gpu_memory_gb"`
	TotalMemoryGB   int       `json:"total_memory_gb"`
	Motherboard     string    `json:"motherboard"`
	BIOSVersion     string    `json:"bios_version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NetworkInterface represents a network interface configuration
type NetworkInterface struct {
	ID            int       `json:"id,omitempty"`
	ServerID      string    `json:"server_id"`
	InterfaceName string    `json:"interface_name"`
	MACAddress    string    `json:"mac_address"`
	InterfaceType string    `json:"interface_type"` // ethernet, wifi, virtual, loopback
	SpeedMbps     int       `json:"speed_mbps"`
	Vendor        string    `json:"vendor"`
	Driver        string    `json:"driver"`
	IsPhysical    bool      `json:"is_physical"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DiskInfo represents disk/storage device information
type DiskInfo struct {
	ID            int       `json:"id,omitempty"`
	ServerID      string    `json:"server_id"`
	DeviceName    string    `json:"device_name"`
	Model         string    `json:"model"`
	SerialNumber  string    `json:"serial_number"`
	SizeGB        int64     `json:"size_gb"`
	DiskType      string    `json:"disk_type"`      // ssd, hdd, nvme, raid
	InterfaceType string    `json:"interface_type"` // sata, nvme, usb
	Filesystem    string    `json:"filesystem"`
	MountPoint    string    `json:"mount_point"`
	IsSystemDisk  bool      `json:"is_system_disk"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CompleteStaticInfo combines all static information for a server
type CompleteStaticInfo struct {
	ServerInfo        *ServerInfo         `json:"server_info"`
	HardwareInfo      *HardwareInfo       `json:"hardware_info"`
	NetworkInterfaces []NetworkInterface  `json:"network_interfaces"`
	DiskInfo          []DiskInfo          `json:"disk_info"`
}

// PostgresStaticDataStorage implements StaticDataStorage using PostgreSQL
type PostgresStaticDataStorage struct {
	db *sql.DB
}

// NewPostgresStaticDataStorage creates a new PostgreSQL-based static data storage
func NewPostgresStaticDataStorage(db *sql.DB) *PostgresStaticDataStorage {
	return &PostgresStaticDataStorage{db: db}
}

// UpsertServerInfo inserts or updates server information
func (s *PostgresStaticDataStorage) UpsertServerInfo(ctx context.Context, info *ServerInfo) error {
	query := `
		INSERT INTO static_data.server_info (
			server_id, hostname, os, os_version, kernel, architecture
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (server_id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			os = EXCLUDED.os,
			os_version = EXCLUDED.os_version,
			kernel = EXCLUDED.kernel,
			architecture = EXCLUDED.architecture,
			updated_at = NOW()
		RETURNING created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query,
		info.ServerID, info.Hostname, info.OS, info.OSVersion, info.Kernel, info.Architecture,
	).Scan(&info.CreatedAt, &info.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert server info: %w", err)
	}
	return nil
}

// GetServerInfo retrieves server information
func (s *PostgresStaticDataStorage) GetServerInfo(ctx context.Context, serverID string) (*ServerInfo, error) {
	query := `
		SELECT server_id, hostname, os, os_version, kernel, architecture, created_at, updated_at
		FROM static_data.server_info
		WHERE server_id = $1`

	info := &ServerInfo{}
	err := s.db.QueryRowContext(ctx, query, serverID).Scan(
		&info.ServerID, &info.Hostname, &info.OS, &info.OSVersion,
		&info.Kernel, &info.Architecture, &info.CreatedAt, &info.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}
	return info, nil
}

// UpsertHardwareInfo inserts or updates hardware information
func (s *PostgresStaticDataStorage) UpsertHardwareInfo(ctx context.Context, info *HardwareInfo) error {
	query := `
		INSERT INTO static_data.hardware_info (
			server_id, cpu_model, cpu_cores, cpu_threads, cpu_frequency_mhz,
			gpu_model, gpu_driver, gpu_memory_gb, total_memory_gb, motherboard, bios_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (server_id) DO UPDATE SET
			cpu_model = EXCLUDED.cpu_model,
			cpu_cores = EXCLUDED.cpu_cores,
			cpu_threads = EXCLUDED.cpu_threads,
			cpu_frequency_mhz = EXCLUDED.cpu_frequency_mhz,
			gpu_model = EXCLUDED.gpu_model,
			gpu_driver = EXCLUDED.gpu_driver,
			gpu_memory_gb = EXCLUDED.gpu_memory_gb,
			total_memory_gb = EXCLUDED.total_memory_gb,
			motherboard = EXCLUDED.motherboard,
			bios_version = EXCLUDED.bios_version,
			updated_at = NOW()
		RETURNING created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query,
		info.ServerID, info.CPUModel, info.CPUCores, info.CPUThreads, info.CPUFrequencyMHz,
		info.GPUModel, info.GPUDriver, info.GPUMemoryGB, info.TotalMemoryGB,
		info.Motherboard, info.BIOSVersion,
	).Scan(&info.CreatedAt, &info.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert hardware info: %w", err)
	}
	return nil
}

// GetHardwareInfo retrieves hardware information
func (s *PostgresStaticDataStorage) GetHardwareInfo(ctx context.Context, serverID string) (*HardwareInfo, error) {
	query := `
		SELECT server_id, cpu_model, cpu_cores, cpu_threads, cpu_frequency_mhz,
			   gpu_model, gpu_driver, gpu_memory_gb, total_memory_gb, motherboard, bios_version,
			   created_at, updated_at
		FROM static_data.hardware_info
		WHERE server_id = $1`

	info := &HardwareInfo{}
	var cpuModel, gpuModel, gpuDriver, motherboard, biosVersion sql.NullString
	var cpuCores, cpuThreads, cpuFreq, gpuMemory, totalMemory sql.NullInt64

	err := s.db.QueryRowContext(ctx, query, serverID).Scan(
		&info.ServerID, &cpuModel, &cpuCores, &cpuThreads, &cpuFreq,
		&gpuModel, &gpuDriver, &gpuMemory, &totalMemory, &motherboard, &biosVersion,
		&info.CreatedAt, &info.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get hardware info: %w", err)
	}

	if cpuModel.Valid {
		info.CPUModel = cpuModel.String
	}
	if cpuCores.Valid {
		info.CPUCores = int(cpuCores.Int64)
	}
	if cpuThreads.Valid {
		info.CPUThreads = int(cpuThreads.Int64)
	}
	if cpuFreq.Valid {
		info.CPUFrequencyMHz = int(cpuFreq.Int64)
	}
	if gpuModel.Valid {
		info.GPUModel = gpuModel.String
	}
	if gpuDriver.Valid {
		info.GPUDriver = gpuDriver.String
	}
	if gpuMemory.Valid {
		info.GPUMemoryGB = int(gpuMemory.Int64)
	}
	if totalMemory.Valid {
		info.TotalMemoryGB = int(totalMemory.Int64)
	}
	if motherboard.Valid {
		info.Motherboard = motherboard.String
	}
	if biosVersion.Valid {
		info.BIOSVersion = biosVersion.String
	}

	return info, nil
}

// UpsertNetworkInterfaces replaces all network interfaces for a server
func (s *PostgresStaticDataStorage) UpsertNetworkInterfaces(ctx context.Context, serverID string, interfaces []NetworkInterface) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing interfaces
	_, err = tx.ExecContext(ctx, "DELETE FROM static_data.network_interfaces WHERE server_id = $1", serverID)
	if err != nil {
		return fmt.Errorf("failed to delete existing interfaces: %w", err)
	}

	// Insert new interfaces
	for _, iface := range interfaces {
		query := `
			INSERT INTO static_data.network_interfaces (
				server_id, interface_name, mac_address, interface_type, speed_mbps, vendor, driver, is_physical
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err = tx.ExecContext(ctx, query,
			serverID, iface.InterfaceName, iface.MACAddress, iface.InterfaceType,
			iface.SpeedMbps, iface.Vendor, iface.Driver, iface.IsPhysical,
		)
		if err != nil {
			return fmt.Errorf("failed to insert interface %s: %w", iface.InterfaceName, err)
		}
	}

	return tx.Commit()
}

// GetNetworkInterfaces retrieves all network interfaces for a server
func (s *PostgresStaticDataStorage) GetNetworkInterfaces(ctx context.Context, serverID string) ([]NetworkInterface, error) {
	query := `
		SELECT id, server_id, interface_name, mac_address, interface_type, speed_mbps, vendor, driver, is_physical, created_at, updated_at
		FROM static_data.network_interfaces
		WHERE server_id = $1
		ORDER BY interface_name`

	rows, err := s.db.QueryContext(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query network interfaces: %w", err)
	}
	defer rows.Close()

	var interfaces []NetworkInterface
	for rows.Next() {
		var iface NetworkInterface
		var macAddr, ifaceType, vendor, driver sql.NullString
		var speedMbps sql.NullInt64

		err := rows.Scan(
			&iface.ID, &iface.ServerID, &iface.InterfaceName, &macAddr, &ifaceType,
			&speedMbps, &vendor, &driver, &iface.IsPhysical, &iface.CreatedAt, &iface.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan interface: %w", err)
		}

		if macAddr.Valid {
			iface.MACAddress = macAddr.String
		}
		if ifaceType.Valid {
			iface.InterfaceType = ifaceType.String
		}
		if speedMbps.Valid {
			iface.SpeedMbps = int(speedMbps.Int64)
		}
		if vendor.Valid {
			iface.Vendor = vendor.String
		}
		if driver.Valid {
			iface.Driver = driver.String
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, rows.Err()
}

// UpsertDiskInfo replaces all disk information for a server
func (s *PostgresStaticDataStorage) UpsertDiskInfo(ctx context.Context, serverID string, disks []DiskInfo) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing disks
	_, err = tx.ExecContext(ctx, "DELETE FROM static_data.disk_info WHERE server_id = $1", serverID)
	if err != nil {
		return fmt.Errorf("failed to delete existing disks: %w", err)
	}

	// Insert new disks
	for _, disk := range disks {
		query := `
			INSERT INTO static_data.disk_info (
				server_id, device_name, model, serial_number, size_gb, disk_type, interface_type, filesystem, mount_point, is_system_disk
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

		_, err = tx.ExecContext(ctx, query,
			serverID, disk.DeviceName, disk.Model, disk.SerialNumber, disk.SizeGB,
			disk.DiskType, disk.InterfaceType, disk.Filesystem, disk.MountPoint, disk.IsSystemDisk,
		)
		if err != nil {
			return fmt.Errorf("failed to insert disk %s: %w", disk.DeviceName, err)
		}
	}

	return tx.Commit()
}

// GetDiskInfo retrieves all disk information for a server
func (s *PostgresStaticDataStorage) GetDiskInfo(ctx context.Context, serverID string) ([]DiskInfo, error) {
	query := `
		SELECT id, server_id, device_name, model, serial_number, size_gb, disk_type, interface_type, filesystem, mount_point, is_system_disk, created_at, updated_at
		FROM static_data.disk_info
		WHERE server_id = $1
		ORDER BY device_name`

	rows, err := s.db.QueryContext(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query disk info: %w", err)
	}
	defer rows.Close()

	var disks []DiskInfo
	for rows.Next() {
		var disk DiskInfo
		var model, serialNum, diskType, ifaceType, fs, mountPoint sql.NullString
		var sizeGB sql.NullInt64

		err := rows.Scan(
			&disk.ID, &disk.ServerID, &disk.DeviceName, &model, &serialNum, &sizeGB,
			&diskType, &ifaceType, &fs, &mountPoint, &disk.IsSystemDisk, &disk.CreatedAt, &disk.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan disk: %w", err)
		}

		if model.Valid {
			disk.Model = model.String
		}
		if serialNum.Valid {
			disk.SerialNumber = serialNum.String
		}
		if sizeGB.Valid {
			disk.SizeGB = sizeGB.Int64
		}
		if diskType.Valid {
			disk.DiskType = diskType.String
		}
		if ifaceType.Valid {
			disk.InterfaceType = ifaceType.String
		}
		if fs.Valid {
			disk.Filesystem = fs.String
		}
		if mountPoint.Valid {
			disk.MountPoint = mountPoint.String
		}

		disks = append(disks, disk)
	}

	return disks, rows.Err()
}

// GetCompleteStaticInfo retrieves all static information for a server
func (s *PostgresStaticDataStorage) GetCompleteStaticInfo(ctx context.Context, serverID string) (*CompleteStaticInfo, error) {
	serverInfo, err := s.GetServerInfo(ctx, serverID)
	if err != nil {
		return nil, err
	}

	hardwareInfo, err := s.GetHardwareInfo(ctx, serverID)
	if err != nil {
		return nil, err
	}

	networkInterfaces, err := s.GetNetworkInterfaces(ctx, serverID)
	if err != nil {
		return nil, err
	}

	diskInfo, err := s.GetDiskInfo(ctx, serverID)
	if err != nil {
		return nil, err
	}

	return &CompleteStaticInfo{
		ServerInfo:        serverInfo,
		HardwareInfo:      hardwareInfo,
		NetworkInterfaces: networkInterfaces,
		DiskInfo:          diskInfo,
	}, nil
}

// UpsertCompleteStaticInfo updates all static information for a server
func (s *PostgresStaticDataStorage) UpsertCompleteStaticInfo(ctx context.Context, serverID string, info *CompleteStaticInfo) error {
	if info.ServerInfo != nil {
		info.ServerInfo.ServerID = serverID
		if err := s.UpsertServerInfo(ctx, info.ServerInfo); err != nil {
			return err
		}
	}

	if info.HardwareInfo != nil {
		info.HardwareInfo.ServerID = serverID
		if err := s.UpsertHardwareInfo(ctx, info.HardwareInfo); err != nil {
			return err
		}
	}

	if info.NetworkInterfaces != nil {
		if err := s.UpsertNetworkInterfaces(ctx, serverID, info.NetworkInterfaces); err != nil {
			return err
		}
	}

	if info.DiskInfo != nil {
		if err := s.UpsertDiskInfo(ctx, serverID, info.DiskInfo); err != nil {
			return err
		}
	}

	return nil
}

// MarshalJSON custom JSON marshaling for CompleteStaticInfo
func (c *CompleteStaticInfo) MarshalJSON() ([]byte, error) {
	type Alias CompleteStaticInfo
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}
