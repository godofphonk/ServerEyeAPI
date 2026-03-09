# Data Structure Separation Plan

## Overview
Разделение данных на статические (отправка 1 раз в сутки) и динамические (отправка каждые 30 секунд).

## Static Data (HTTP POST once per day)

### Server Info
```json
{
  "hostname": "gospodin-A620M-Pro-RS",
  "os": "Ubuntu",
  "os_version": "25.10",
  "kernel": "6.17.0-14-generic",
  "architecture": "x86_64"
}
```

### Hardware Info
```json
{
  "cpu_model": "AMD Ryzen 7 8700F 8-Core Processor",
  "cpu_cores": 8,
  "cpu_threads": 16,
  "cpu_frequency_mhz": 4799.55,
  "gpu_model": "NVIDIA GeForce RTX 4060",
  "gpu_driver": "590.48.01",
  "gpu_memory_gb": 8.0,
  "total_memory_gb": 32.0,
  "motherboard": "ASUS ROG",
  "bios_version": "1.2.3"
}
```

### Network Interfaces (Static)
```json
{
  "network_interfaces": [
    {
      "interface_name": "enp111s0",
      "mac_address": "00:11:22:33:44:55",
      "interface_type": "ethernet",
      "speed_mbps": 1000,
      "vendor": "Intel",
      "driver": "e1000e",
      "is_physical": true
    },
    {
      "interface_name": "docker0",
      "mac_address": "02:42:ab:cd:ef:gh",
      "interface_type": "virtual",
      "speed_mbps": 0,
      "vendor": "Docker",
      "driver": "bridge",
      "is_physical": false
    }
  ]
}
```

### Disk Info (Static)
```json
{
  "disk_info": [
    {
      "device_name": "/dev/nvme0n1p2",
      "filesystem_type": "ext4",
      "mount_point": "/",
      "total_gb": 553
    },
    {
      "device_name": "/dev/nvme0n1p1",
      "filesystem_type": "vfat",
      "mount_point": "/boot/efi",
      "total_gb": 0.5
    }
  ]
}
```

### Temperature Sensors (Static)
```json
{
  "temperature_sensors": [
    {
      "sensor_type": "cpu",
      "sensor_name": "CPU Package"
    },
    {
      "sensor_type": "gpu",
      "sensor_name": "NVIDIA GPU"
    },
    {
      "sensor_type": "storage",
      "device": "/dev/nvme0n1",
      "storage_type": "NVMe"
    }
  ]
}
```

---

## Dynamic Data (WebSocket every 30 seconds)

### Simplified Structure
```json
{
  "type": "metrics",
  "server_id": "srv_xxx",
  "data": {
    "cpu_usage": {
      "usage_total": 20.10,
      "usage_user": 17.38,
      "usage_system": 2.52,
      "usage_idle": 79.89,
      "load_average": {
        "load_1min": 2.21,
        "load_5min": 2.05,
        "load_15min": 2.23
      },
      "frequency_mhz": 2140.709
    },
    "memory": {
      "total_gb": 29.97,
      "used_gb": 19.40,
      "available_gb": 10.57,
      "free_gb": 0.55,
      "buffers_gb": 0.52,
      "cached_gb": 10.82,
      "used_percent": 64.72
    },
    "disks": [
      {
        "mount_point": "/",
        "device_name": "/dev/nvme0n1p2",
        "used_gb": 418,
        "free_gb": 107,
        "used_percent": 80
      },
      {
        "mount_point": "/boot/efi",
        "device_name": "/dev/nvme0n1p1",
        "used_gb": 0.04,
        "free_gb": 0.46,
        "used_percent": 4
      }
    ],
    "network": {
      "interfaces": [
        {
          "name": "enp111s0",
          "rx_bytes": 1256214555,
          "tx_bytes": 100710535,
          "rx_packets": 1006104,
          "tx_packets": 397614,
          "rx_speed_mbps": 0.125,
          "tx_speed_mbps": 0.212,
          "status": "up"
        }
      ],
      "total_rx_mbps": 0.343,
      "total_tx_mbps": 0.621
    },
    "temperature": {
      "cpu": 32,
      "gpu": 49,
      "storage": [
        {
          "device": "/dev/nvme0n1",
          "temperature": 29.85
        }
      ],
      "highest": 49
    },
    "system": {
      "processes_total": 645,
      "processes_running": 1,
      "processes_sleeping": 644,
      "uptime_seconds": 9578
    },
    "timestamp": "2026-03-09T15:57:09.428681405+05:00"
  },
  "timestamp": 1773053829
}
```

---

## Changes Summary

### Removed from Dynamic Metrics
- ❌ Top-level `cpu`, `memory`, `disk`, `network`, `temperature` (unclear aggregated values)
- ❌ `system.os`, `system.architecture`, `system.kernel`, `system.hostname` (duplicates, moved to static)
- ❌ `uptime_human` (can be calculated on frontend)
- ❌ `temperature_unit` (always celsius)
- ❌ `system_temperature` (always 0)

### Added to Static Data
- ✅ `cpu_cores` from `cpu_usage.cores`
- ✅ `mac_address` for network interfaces (needs to be collected by agent)
- ✅ `filesystem_type` for disks (ext4, vfat, etc.)
- ✅ Temperature sensors list

### Renamed/Restructured
- `disk_details` → `disks` (simpler name)
- `network_details.interfaces` → `network.interfaces`
- `temperature_details` → `temperature`
- `memory_details` → `memory`
- `cpu_usage.frequency` → `cpu_usage.frequency_mhz` (explicit unit)

---

## Database Schema Changes

### TimescaleDB (Dynamic Metrics)
- Keep existing `metrics` hypertable
- Add columns for new fields if needed
- Remove unused columns

### Static PostgreSQL
- Update `hardware_info` table (add cores if not exists)
- Update `network_interfaces` table (ensure mac_address exists)
- Update `disk_info` table (add filesystem_type)
- Create `temperature_sensors` table (new)

---

## Implementation Steps

1. ✅ Update Go models for dynamic metrics
2. ✅ Update Go models for static data
3. ✅ Create/update database migrations
4. ✅ Update WebSocket handler
5. ✅ Update HTTP static info handler
6. ✅ Test with real agent data
7. ✅ Update agent to send new structure
