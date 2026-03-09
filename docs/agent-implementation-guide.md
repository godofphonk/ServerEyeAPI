# Agent Implementation Guide

## Data Separation Strategy

### Static Data (HTTP POST - once per day)
**Endpoint:** `POST /api/servers/by-key/{server_key}/static-info`

**What to send:**
- Server basic info (hostname, OS, kernel, architecture)
- Hardware specifications (CPU model, cores, threads, GPU, memory total)
- Network interfaces list (name, MAC, type, vendor, driver)
- Disk devices list (device name, filesystem type, mount point, total size)
- Temperature sensors available

**When to send:**
- Once per day at consistent time (e.g., 03:00 AM)
- On agent startup
- When hardware changes detected

### Dynamic Data (WebSocket - every 30 seconds)
**Endpoint:** WebSocket `/ws`

**What to send:**
- CPU usage (total, user, system, idle, load average, current frequency)
- Memory usage (used, available, free, buffers, cached, percent)
- Disk usage per mount point (used, free, percent)
- Network statistics per interface (bytes, packets, speed, status)
- Temperature readings (CPU, GPU, storage devices)
- System metrics (processes, uptime)

## Required Changes in Agent

### 1. Remove from Dynamic Metrics
```json
{
  "cpu": 20.10,        // ❌ Remove - unclear aggregated value
  "memory": 64.72,     // ❌ Remove - unclear aggregated value
  "disk": 80,          // ❌ Remove - unclear aggregated value
  "network": 0.96,     // ❌ Remove - unclear aggregated value
  "temperature": 0     // ❌ Remove - always 0
}
```

### 2. Move to Static Data
```json
{
  "cpu_usage": {
    "cores": 16        // ✅ Move to static hardware_info
  },
  "system_details": {
    "hostname": "...", // ✅ Move to static server_info
    "os": "...",       // ✅ Move to static server_info
    "kernel": "...",   // ✅ Move to static server_info
    "architecture": "..." // ✅ Move to static server_info
  }
}
```

### 3. Rename Fields
```json
{
  "cpu_usage": {
    "frequency": 2140.709  // ✅ Rename to "frequency_mhz"
  },
  "disk_details": [...],   // ✅ Rename to "disks"
  "network_details": {...}, // ✅ Rename to "network"
  "temperature_details": {...}, // ✅ Rename to "temperature"
  "memory_details": {...}  // ✅ Rename to "memory"
}
```

### 4. Remove Unnecessary Fields
```json
{
  "system_details": {
    "uptime_human": "...",      // ❌ Remove - can be calculated on frontend
    "boot_time": "..."          // ❌ Remove - not needed
  },
  "temperature_details": {
    "temperature_unit": "celsius", // ❌ Remove - always celsius
    "system_temperature": 0        // ❌ Remove - always 0
  }
}
```

### 5. Add Missing Fields to Static Data
```json
{
  "network_interfaces": [
    {
      "interface_name": "enp111s0",
      "mac_address": "00:11:22:33:44:55", // ✅ ADD - collect MAC address
      "interface_type": "ethernet",        // ✅ ADD - determine type
      "speed_mbps": 1000,                 // ✅ ADD - link speed
      "vendor": "Intel",                  // ✅ ADD if available
      "driver": "e1000e",                 // ✅ ADD if available
      "is_physical": true                 // ✅ ADD - physical vs virtual
    }
  ],
  "disk_info": [
    {
      "device_name": "/dev/nvme0n1p2",
      "filesystem_type": "ext4",    // ✅ ADD - filesystem type
      "mount_point": "/",
      "total_gb": 553
    }
  ]
}
```

## New Dynamic Metrics Structure

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
    "timestamp": "2026-03-09T15:57:09Z"
  },
  "timestamp": 1773053829
}
```

## Implementation Checklist

### Agent Side
- [ ] Collect MAC addresses for network interfaces
- [ ] Determine interface type (ethernet/virtual/loopback)
- [ ] Collect link speed for interfaces
- [ ] Collect vendor and driver info if available
- [ ] Collect filesystem type for disks
- [ ] Implement static data HTTP POST (once per day)
- [ ] Update dynamic metrics WebSocket structure
- [ ] Remove deprecated fields
- [ ] Rename fields as specified
- [ ] Test both static and dynamic data sending

### API Side
- [x] Create new metrics models (MetricsV2)
- [ ] Update database schemas
- [ ] Update WebSocket handler for new structure
- [ ] Ensure backward compatibility during transition
- [ ] Update documentation
- [ ] Test with new agent data
