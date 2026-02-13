# Test Servers

This directory contains test servers for demonstrating the ServerEyeAPI multi-tier metrics system.

## Structure

```
test_servers/
├── metrics_server/
│   └── main.go          - Metrics API endpoints (port 8083)
├── commands_server/
│   └── main.go          - Management commands (port 8084)
└── README.md
```

## Usage

### Run Metrics Test Server

```bash
cd test_servers/metrics_server
go run main.go
```

### Run Commands Test Server

```bash
cd test_servers/commands_server
go run main.go
```

### Run Demo

From the root directory:

```bash
./demo_system.sh
```

## Notes

These servers use mock data for demonstration purposes. They don't require a database connection.
