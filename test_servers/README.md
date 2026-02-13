# Test Servers

This directory contains test servers for demonstrating the ServerEyeAPI multi-tier metrics system.

## Files

- `test_metrics_endpoints.go` - Test server for metrics API endpoints (port 8083)
- `test_commands.go` - Test server for management commands (port 8084)

## Usage

### Run Metrics Test Server

```bash
cd test_servers
go build -o test_metrics test_metrics_endpoints.go
./test_metrics
```

### Run Commands Test Server

```bash
cd test_servers
go build -o test_commands test_commands.go
./test_commands
```

### Run Demo

From the root directory:

```bash
./demo_system.sh
```

## Notes

These servers use mock data for demonstration purposes. They don't require a database connection.
