package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// StaticDataStorage handles static/persistent server information
type StaticDataStorage interface {
	// Server Info
	UpsertServerInfo(ctx context.Context, info *ServerInfo) error
	GetServerInfo(ctx context.Context, serverID string) (*ServerInfo, error)

	// Hardware Info
	UpsertHardwareInfo(ctx context.Context, info *HardwareInfo) error
	GetHardwareInfo(ctx context.Context, serverID string) (*HardwareInfo, error)

	// Motherboard Info
	UpsertMotherboardInfo(ctx context.Context, info *MotherboardInfo) error
	GetMotherboardInfo(ctx context.Context, serverID string) (*MotherboardInfo, error)

	// Memory Modules
	UpsertMemoryModules(ctx context.Context, serverID string, modules []MemoryModule) error
	GetMemoryModules(ctx context.Context, serverID string) ([]MemoryModule, error)

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

// HardwareInfo represents server hardware specifications
type HardwareInfo struct {
	ServerID        string    `json:"server_id"`
	CPUModel        string    `json:"cpu_model"`
	CPUCores        int       `json:"cpu_cores"`
	CPUThreads      int       `json:"cpu_threads"`
	CPUFrequencyMHz float64   `json:"cpu_frequency_mhz"`
	GPUModel        string    `json:"gpu_model"`
	GPUDriver       string    `json:"gpu_driver"`
	GPUMemoryGB     int       `json:"gpu_memory_gb"`
	TotalMemoryGB   float64   `json:"total_memory_gb"`
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

// MemoryModule represents individual memory module information
type MemoryModule struct {
	ID           int       `json:"id,omitempty"`
	ServerID     string    `json:"server_id"`
	SlotName     string    `json:"slot_name"`
	SizeGB       int       `json:"size_gb"`
	MemoryType   string    `json:"memory_type"` // DDR3, DDR4, DDR5
	FrequencyMHz int       `json:"frequency_mhz"`
	Manufacturer string    `json:"manufacturer"`
	PartNumber   string    `json:"part_number"`
	SpeedMTs     int       `json:"speed_mts"`  // For DDR5 (MT/s)
	Voltage      float64   `json:"voltage"`    // Memory voltage
	Timings      string    `json:"timings"`    // CAS timings
	ECC          bool      `json:"ecc"`        // ECC memory
	Registered   bool      `json:"registered"` // Registered memory
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MotherboardInfo represents extended motherboard information
type MotherboardInfo struct {
	ServerID             string    `json:"server_id"`
	Manufacturer         string    `json:"manufacturer,omitempty"`
	Model                string    `json:"model,omitempty"`
	Chipset              string    `json:"chipset,omitempty"`
	BIOSVersion          string    `json:"bios_version,omitempty"`
	BIOSDate             time.Time `json:"bios_date,omitempty"`
	BIOSVendor           string    `json:"bios_vendor,omitempty"`
	FormFactor           string    `json:"form_factor,omitempty"` // ATX, Micro-ATX, Mini-ITX
	MaxMemoryGB          int       `json:"max_memory_gb,omitempty"`
	MemorySlots          int       `json:"memory_slots,omitempty"`
	SupportedMemoryTypes []string  `json:"supported_memory_types,omitempty"` // ['DDR4', 'DDR5']
	OnboardVideo         bool      `json:"onboard_video,omitempty"`
	OnboardAudio         bool      `json:"onboard_audio,omitempty"`
	OnboardNetwork       bool      `json:"onboard_network,omitempty"`
	SATAPorts            int       `json:"sata_ports,omitempty"`
	SATASpeed            string    `json:"sata_speed,omitempty"` // SATA 3.0, SATA 6.0
	M2Slots              int       `json:"m2_slots,omitempty"`
	PCIeSlots            []string  `json:"pcie_slots,omitempty"` // ['x16', 'x8', 'x4']
	USBPortsTotal        int       `json:"usb_ports_total,omitempty"`
	USBPorts20           int       `json:"usb_ports_2_0,omitempty"`
	USBPorts30           int       `json:"usb_ports_3_0,omitempty"`
	USBPortsC            int       `json:"usb_ports_c,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CompleteStaticInfo combines all static information for a server
type CompleteStaticInfo struct {
	ServerInfo        *ServerInfo        `json:"server_info,omitempty"`
	HardwareInfo      *HardwareInfo      `json:"hardware_info,omitempty"`
	MotherboardInfo   *MotherboardInfo   `json:"motherboard_info,omitempty"`
	MemoryModules     []MemoryModule     `json:"memory_modules,omitempty"`
	NetworkInterfaces []NetworkInterface `json:"network_interfaces,omitempty"`
	DiskInfo          []DiskInfo         `json:"disk_info,omitempty"`
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
			gpu_model, gpu_driver, gpu_memory_gb, total_memory_gb
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (server_id) DO UPDATE SET
			cpu_model = EXCLUDED.cpu_model,
			cpu_cores = EXCLUDED.cpu_cores,
			cpu_threads = EXCLUDED.cpu_threads,
			cpu_frequency_mhz = EXCLUDED.cpu_frequency_mhz,
			gpu_model = EXCLUDED.gpu_model,
			gpu_driver = EXCLUDED.gpu_driver,
			gpu_memory_gb = EXCLUDED.gpu_memory_gb,
			total_memory_gb = EXCLUDED.total_memory_gb,
			updated_at = NOW()
		RETURNING created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query,
		info.ServerID, info.CPUModel, info.CPUCores, info.CPUThreads, info.CPUFrequencyMHz,
		info.GPUModel, info.GPUDriver, info.GPUMemoryGB, info.TotalMemoryGB,
	).Scan(&info.CreatedAt, &info.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert hardware info: %w", err)
	}
	return nil
}
func (s *PostgresStaticDataStorage) GetHardwareInfo(ctx context.Context, serverID string) (*HardwareInfo, error) {
	query := `
		SELECT server_id, cpu_model, cpu_cores, cpu_threads, cpu_frequency_mhz,
			   gpu_model, gpu_driver, gpu_memory_gb, total_memory_gb,
			   created_at, updated_at
		FROM static_data.hardware_info
		WHERE server_id = $1`

	info := &HardwareInfo{}
	var cpuModel, gpuModel, gpuDriver sql.NullString
	var cpuCores, cpuThreads, gpuMemoryGB sql.NullInt64
	var cpuFreq, totalMemory sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, serverID).Scan(
		&info.ServerID, &cpuModel, &cpuCores, &cpuThreads, &cpuFreq,
		&gpuModel, &gpuDriver, &gpuMemoryGB, &totalMemory,
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
	if gpuModel.Valid {
		info.GPUModel = gpuModel.String
	}
	if gpuDriver.Valid {
		info.GPUDriver = gpuDriver.String
	}
	if cpuCores.Valid {
		info.CPUCores = int(cpuCores.Int64)
	}
	if cpuThreads.Valid {
		info.CPUThreads = int(cpuThreads.Int64)
	}
	if cpuFreq.Valid {
		info.CPUFrequencyMHz = cpuFreq.Float64
	}
	if gpuMemoryGB.Valid {
		info.GPUMemoryGB = int(gpuMemoryGB.Int64)
	}
	if totalMemory.Valid {
		info.TotalMemoryGB = totalMemory.Float64
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
	result := &CompleteStaticInfo{}

	serverInfo, err := s.GetServerInfo(ctx, serverID)
	if err == nil && serverInfo != nil {
		result.ServerInfo = serverInfo
	}

	hardwareInfo, err := s.GetHardwareInfo(ctx, serverID)
	if err == nil && hardwareInfo != nil {
		result.HardwareInfo = hardwareInfo
	}

	motherboardInfo, err := s.GetMotherboardInfo(ctx, serverID)
	if err == nil && motherboardInfo != nil {
		// Only include motherboard info if it has meaningful data
		if motherboardInfo.Manufacturer != "" || motherboardInfo.Model != "" {
			// Create a filtered version without empty/zero fields
			filtered := &MotherboardInfo{
				ServerID:  motherboardInfo.ServerID,
				CreatedAt: motherboardInfo.CreatedAt,
				UpdatedAt: motherboardInfo.UpdatedAt,
			}

			if motherboardInfo.Manufacturer != "" {
				filtered.Manufacturer = motherboardInfo.Manufacturer
			}
			if motherboardInfo.Model != "" {
				filtered.Model = motherboardInfo.Model
			}
			if motherboardInfo.Chipset != "" && motherboardInfo.Chipset != "Unknown" {
				filtered.Chipset = motherboardInfo.Chipset
			}
			if motherboardInfo.BIOSVersion != "" {
				filtered.BIOSVersion = motherboardInfo.BIOSVersion
			}
			if !motherboardInfo.BIOSDate.IsZero() && motherboardInfo.BIOSDate.Year() > 1 {
				filtered.BIOSDate = motherboardInfo.BIOSDate
			}
			if motherboardInfo.BIOSVendor != "" {
				filtered.BIOSVendor = motherboardInfo.BIOSVendor
			}
			if motherboardInfo.FormFactor != "" {
				filtered.FormFactor = motherboardInfo.FormFactor
			}
			if motherboardInfo.MaxMemoryGB > 0 {
				filtered.MaxMemoryGB = motherboardInfo.MaxMemoryGB
			}
			if motherboardInfo.MemorySlots > 0 {
				filtered.MemorySlots = motherboardInfo.MemorySlots
			}
			if len(motherboardInfo.SupportedMemoryTypes) > 0 {
				filtered.SupportedMemoryTypes = motherboardInfo.SupportedMemoryTypes
			}
			if motherboardInfo.OnboardVideo {
				filtered.OnboardVideo = motherboardInfo.OnboardVideo
			}
			if motherboardInfo.OnboardAudio {
				filtered.OnboardAudio = motherboardInfo.OnboardAudio
			}
			if motherboardInfo.OnboardNetwork {
				filtered.OnboardNetwork = motherboardInfo.OnboardNetwork
			}
			if motherboardInfo.SATAPorts > 0 {
				filtered.SATAPorts = motherboardInfo.SATAPorts
			}
			if motherboardInfo.SATASpeed != "" {
				filtered.SATASpeed = motherboardInfo.SATASpeed
			}
			if motherboardInfo.M2Slots > 0 {
				filtered.M2Slots = motherboardInfo.M2Slots
			}
			if len(motherboardInfo.PCIeSlots) > 0 {
				filtered.PCIeSlots = motherboardInfo.PCIeSlots
			}
			if motherboardInfo.USBPortsTotal > 0 {
				filtered.USBPortsTotal = motherboardInfo.USBPortsTotal
			}
			if motherboardInfo.USBPorts20 > 0 {
				filtered.USBPorts20 = motherboardInfo.USBPorts20
			}
			if motherboardInfo.USBPorts30 > 0 {
				filtered.USBPorts30 = motherboardInfo.USBPorts30
			}
			if motherboardInfo.USBPortsC > 0 {
				filtered.USBPortsC = motherboardInfo.USBPortsC
			}

			result.MotherboardInfo = filtered
		}
	}

	memoryModules, err := s.GetMemoryModules(ctx, serverID)
	if err == nil && len(memoryModules) > 0 {
		result.MemoryModules = memoryModules
	}

	networkInterfaces, err := s.GetNetworkInterfaces(ctx, serverID)
	if err == nil && len(networkInterfaces) > 0 {
		result.NetworkInterfaces = networkInterfaces
	}

	diskInfo, err := s.GetDiskInfo(ctx, serverID)
	if err == nil && len(diskInfo) > 0 {
		result.DiskInfo = diskInfo
	}

	return result, nil
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

	if info.MotherboardInfo != nil {
		info.MotherboardInfo.ServerID = serverID
		if err := s.UpsertMotherboardInfo(ctx, info.MotherboardInfo); err != nil {
			return err
		}
	}

	if info.MemoryModules != nil {
		if err := s.UpsertMemoryModules(ctx, serverID, info.MemoryModules); err != nil {
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

// UpsertMotherboardInfo inserts or updates motherboard information
func (s *PostgresStaticDataStorage) UpsertMotherboardInfo(ctx context.Context, info *MotherboardInfo) error {
	query := `
		INSERT INTO static_data.motherboard_info (
			server_id, manufacturer, model, chipset, bios_version, bios_date, bios_vendor,
			form_factor, max_memory_gb, memory_slots, supported_memory_types,
			onboard_video, onboard_audio, onboard_network, sata_ports, sata_speed,
			m2_slots, pcie_slots, usb_ports_total, usb_ports_2_0, usb_ports_3_0, usb_ports_c
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		ON CONFLICT (server_id) DO UPDATE SET
			manufacturer = EXCLUDED.manufacturer,
			model = EXCLUDED.model,
			chipset = EXCLUDED.chipset,
			bios_version = EXCLUDED.bios_version,
			bios_date = EXCLUDED.bios_date,
			bios_vendor = EXCLUDED.bios_vendor,
			form_factor = EXCLUDED.form_factor,
			max_memory_gb = EXCLUDED.max_memory_gb,
			memory_slots = EXCLUDED.memory_slots,
			supported_memory_types = EXCLUDED.supported_memory_types,
			onboard_video = EXCLUDED.onboard_video,
			onboard_audio = EXCLUDED.onboard_audio,
			onboard_network = EXCLUDED.onboard_network,
			sata_ports = EXCLUDED.sata_ports,
			sata_speed = EXCLUDED.sata_speed,
			m2_slots = EXCLUDED.m2_slots,
			pcie_slots = EXCLUDED.pcie_slots,
			usb_ports_total = EXCLUDED.usb_ports_total,
			usb_ports_2_0 = EXCLUDED.usb_ports_2_0,
			usb_ports_3_0 = EXCLUDED.usb_ports_3_0,
			usb_ports_c = EXCLUDED.usb_ports_c,
			updated_at = NOW()
		RETURNING created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query,
		info.ServerID, info.Manufacturer, info.Model, info.Chipset, info.BIOSVersion, info.BIOSDate, info.BIOSVendor,
		info.FormFactor, info.MaxMemoryGB, info.MemorySlots, info.SupportedMemoryTypes,
		info.OnboardVideo, info.OnboardAudio, info.OnboardNetwork, info.SATAPorts, info.SATASpeed,
		info.M2Slots, info.PCIeSlots, info.USBPortsTotal, info.USBPorts20, info.USBPorts30, info.USBPortsC,
	).Scan(&info.CreatedAt, &info.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert motherboard info: %w", err)
	}
	return nil
}

// GetMotherboardInfo retrieves motherboard information
func (s *PostgresStaticDataStorage) GetMotherboardInfo(ctx context.Context, serverID string) (*MotherboardInfo, error) {
	query := `
		SELECT server_id, manufacturer, model, chipset, bios_version, bios_date, bios_vendor,
			   form_factor, max_memory_gb, memory_slots, supported_memory_types,
			   onboard_video, onboard_audio, onboard_network, sata_ports, sata_speed,
			   m2_slots, pcie_slots, usb_ports_total, usb_ports_2_0, usb_ports_3_0, usb_ports_c,
			   created_at, updated_at
		FROM static_data.motherboard_info
		WHERE server_id = $1`

	info := &MotherboardInfo{}
	var manufacturer, model, chipset, biosVersion, biosVendor, formFactor, sataSpeed sql.NullString
	var maxMemory, memorySlots, sataPorts, m2Slots, usbTotal, usb20, usb30, usbc sql.NullInt64
	var supportedMemoryTypes, pcieSlots pq.StringArray

	err := s.db.QueryRowContext(ctx, query, serverID).Scan(
		&info.ServerID, &manufacturer, &model, &chipset, &biosVersion, &info.BIOSDate, &biosVendor,
		&formFactor, &maxMemory, &memorySlots, &supportedMemoryTypes,
		&info.OnboardVideo, &info.OnboardAudio, &info.OnboardNetwork, &sataPorts, &sataSpeed,
		&m2Slots, &pcieSlots, &usbTotal, &usb20, &usb30, &usbc,
		&info.CreatedAt, &info.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get motherboard info: %w", err)
	}

	if manufacturer.Valid {
		info.Manufacturer = manufacturer.String
	}
	if model.Valid {
		info.Model = model.String
	}
	if chipset.Valid {
		info.Chipset = chipset.String
	}
	if biosVersion.Valid {
		info.BIOSVersion = biosVersion.String
	}
	if biosVendor.Valid {
		info.BIOSVendor = biosVendor.String
	}
	if formFactor.Valid {
		info.FormFactor = formFactor.String
	}
	if maxMemory.Valid {
		info.MaxMemoryGB = int(maxMemory.Int64)
	}
	if memorySlots.Valid {
		info.MemorySlots = int(memorySlots.Int64)
	}
	if sataPorts.Valid {
		info.SATAPorts = int(sataPorts.Int64)
	}
	if m2Slots.Valid {
		info.M2Slots = int(m2Slots.Int64)
	}
	if usbTotal.Valid {
		info.USBPortsTotal = int(usbTotal.Int64)
	}
	if usb20.Valid {
		info.USBPorts20 = int(usb20.Int64)
	}
	if usb30.Valid {
		info.USBPorts30 = int(usb30.Int64)
	}
	if usbc.Valid {
		info.USBPortsC = int(usbc.Int64)
	}
	if sataSpeed.Valid {
		info.SATASpeed = sataSpeed.String
	}
	info.SupportedMemoryTypes = []string(supportedMemoryTypes)
	info.PCIeSlots = []string(pcieSlots)

	return info, nil
}

// UpsertMemoryModules replaces all memory modules for a server
func (s *PostgresStaticDataStorage) UpsertMemoryModules(ctx context.Context, serverID string, modules []MemoryModule) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing modules
	_, err = tx.ExecContext(ctx, "DELETE FROM static_data.memory_modules WHERE server_id = $1", serverID)
	if err != nil {
		return fmt.Errorf("failed to delete existing memory modules: %w", err)
	}

	// Insert new modules
	for _, module := range modules {
		query := `
			INSERT INTO static_data.memory_modules (
				server_id, slot_name, size_gb, memory_type, frequency_mhz, manufacturer,
				part_number, speed_mts, voltage, timings, ecc, registered
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

		_, err = tx.ExecContext(ctx, query,
			serverID, module.SlotName, module.SizeGB, module.MemoryType, module.FrequencyMHz,
			module.Manufacturer, module.PartNumber, module.SpeedMTs, module.Voltage,
			module.Timings, module.ECC, module.Registered,
		)
		if err != nil {
			return fmt.Errorf("failed to insert memory module %s: %w", module.SlotName, err)
		}
	}

	return tx.Commit()
}

// GetMemoryModules retrieves all memory modules for a server
func (s *PostgresStaticDataStorage) GetMemoryModules(ctx context.Context, serverID string) ([]MemoryModule, error) {
	query := `
		SELECT id, server_id, slot_name, size_gb, memory_type, frequency_mhz, manufacturer,
			   part_number, speed_mts, voltage, timings, ecc, registered, created_at, updated_at
		FROM static_data.memory_modules
		WHERE server_id = $1
		ORDER BY slot_name`

	rows, err := s.db.QueryContext(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory modules: %w", err)
	}
	defer rows.Close()

	var modules []MemoryModule
	for rows.Next() {
		var module MemoryModule
		var memoryType, manufacturer, partNumber, timings sql.NullString
		var frequencyMHz, speedMTs sql.NullInt64
		var voltage sql.NullFloat64

		err := rows.Scan(
			&module.ID, &module.ServerID, &module.SlotName, &module.SizeGB, &memoryType,
			&frequencyMHz, &manufacturer, &partNumber, &speedMTs, &voltage, &timings,
			&module.ECC, &module.Registered, &module.CreatedAt, &module.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory module: %w", err)
		}

		if memoryType.Valid {
			module.MemoryType = memoryType.String
		}
		if manufacturer.Valid {
			module.Manufacturer = manufacturer.String
		}
		if partNumber.Valid {
			module.PartNumber = partNumber.String
		}
		if frequencyMHz.Valid {
			module.FrequencyMHz = int(frequencyMHz.Int64)
		}
		if speedMTs.Valid {
			module.SpeedMTs = int(speedMTs.Int64)
		}
		if voltage.Valid {
			module.Voltage = voltage.Float64
		}
		if timings.Valid {
			module.Timings = timings.String
		}

		modules = append(modules, module)
	}

	return modules, rows.Err()
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
