# ServerEye API Makefile

.PHONY: build build-api test test-coverage test-coverage-threshold test-integration test-all install-coverage-tools clean docker-build docker-run docker-stop

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Test targets
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	$(GOTEST) -v -race -coverprofile=coverage/coverage.out ./...
	@echo "Generating coverage report..."
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated: coverage/coverage.html"

test-coverage-threshold:
	@echo "Running tests with coverage threshold check..."
	@mkdir -p coverage
	$(GOTEST) -v -race -coverprofile=coverage/coverage.out ./...
	@COVERAGE=$$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$COVERAGE%"; \
	if (( $$(echo "$$COVERAGE < 80" | bc -l) )); then \
		echo "❌ Coverage is below 80%"; \
		exit 1; \
	else \
		echo "✅ Coverage threshold met!"; \
	fi

test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

test-all: test test-coverage

# Coverage tools
install-coverage-tools:
	@echo "Installing coverage tools..."
	go install github.com/wadey/gocovmerge@latest
	go install github.com/axw/gocov/gocov@latest

# Binary name
API_BINARY=servereye-api

# Build directories
BUILD_DIR=build

# Default target
all: build

# Build API only
build: build-api

# Version and build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS = -X github.com/godofphonk/ServerEyeAPI/internal/version.Version=$(VERSION) \
	-X github.com/godofphonk/ServerEyeAPI/internal/version.BuildTime=$(BUILD_DATE) \
	-X github.com/godofphonk/ServerEyeAPI/internal/version.GitCommit=$(GIT_COMMIT)

# Build API
build-api:
	@echo "Building API $(VERSION)..."
	@echo "Generating Wire dependencies..."
	go generate ./internal/wire
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(API_BINARY) ./cmd/api
	@echo "✅ API built: $(BUILD_DIR)/$(API_BINARY)"

# Run API locally
run:
	@echo "Running API..."
	@echo "Generating Wire dependencies..."
	go generate ./internal/wire
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(API_BINARY) ./cmd/api
	./$(BUILD_DIR)/$(API_BINARY)

# Run all tests
test:
	@echo "Running all tests..."
	$(GOTEST) -v -short ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -short -coverprofile=coverage.txt -covermode=atomic ./...
	@echo ""
	@echo "📊 Coverage Report:"
	go tool cover -func=coverage.txt | grep -E "internal/agent|^total:"

# Test specific module
test-agent:
	@echo "Testing internal/agent..."
	$(GOTEST) -v -short -cover ./internal/agent/...

test-pkg:
	@echo "Testing pkg modules..."
	$(GOTEST) -v -short -cover ./pkg/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Copy Linux binaries to downloads folder
downloads: release
	@echo "Copying Linux binaries to downloads folder..."
	@mkdir -p downloads
	@cp $(BUILD_DIR)/$(AGENT_BINARY)-linux-amd64 downloads/
	@cp $(BUILD_DIR)/$(AGENT_BINARY)-linux-arm64 downloads/
	@chmod +x downloads/$(AGENT_BINARY)-linux-amd64
	@chmod +x downloads/$(AGENT_BINARY)-linux-arm64
	@echo "✅ Linux binaries copied to downloads/"
	@ls -la downloads/

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Docker targets
docker-build:
	docker build -t servereye-api:latest .

docker-run:
	docker run -p 8080:8080 --env-file .env servereye-api:latest

docker-push:
	docker tag servereye-api:latest ghcr.io/godofphonk/ServerEyeAPI:latest
	docker push ghcr.io/godofphonk/ServerEyeAPI:latest

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-compose-logs:
	docker-compose logs -f

# Start services with Docker Compose
docker-up:
	@echo "Starting services..."
	cd deployments && docker-compose up -d

# Stop services
docker-down:
	@echo "Stopping services..."
	cd deployments && docker-compose down

# View logs
docker-logs:
	cd deployments && docker-compose logs -f

# Install agent (Linux only)
install-agent: build-agent
	@echo "Installing agent..."
	sudo cp $(BUILD_DIR)/$(AGENT_BINARY) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(AGENT_BINARY)
	@echo "Agent installed to /usr/local/bin/$(AGENT_BINARY)"
	@echo "Run 'sudo $(AGENT_BINARY) --install' to complete setup"

# Development targets
dev-agent:
	@echo "Running agent in development mode..."
	$(GOCMD) run ./cmd/agent --log-level=debug

# Generate mocks (requires mockgen)
mocks:
	@echo "Generating mocks and Wire dependencies..."
	go generate ./...

# Security scan (requires gosec)
security:
	@echo "Running security scan..."
	gosec ./...

# Check for vulnerabilities (requires govulncheck)
vuln-check:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

# Release build with optimizations and version
RELEASE_LDFLAGS = -w -s \
	-X github.com/servereye/servereye/internal/version.Version=$(VERSION) \
	-X github.com/servereye/servereye/internal/version.BuildDate=$(BUILD_DATE) \
	-X github.com/servereye/servereye/internal/version.GitCommit=$(GIT_COMMIT)

release: clean
	@echo "Building release binaries for version $(VERSION)..."
	@echo "Build date: $(BUILD_DATE)"
	@echo "Git commit: $(GIT_COMMIT)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(RELEASE_LDFLAGS)" -o $(BUILD_DIR)/$(AGENT_BINARY)-linux-amd64 ./cmd/agent
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="$(RELEASE_LDFLAGS)" -o $(BUILD_DIR)/$(AGENT_BINARY)-linux-arm64 ./cmd/agent
	@echo "✅ Release build complete!"
	@$(BUILD_DIR)/$(AGENT_BINARY)-linux-amd64 --version

# Показать текущую версию
version:
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Git Commit: $(GIT_COMMIT)"

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build:"
	@echo "  build         - Build agent"
	@echo "  build-agent   - Build agent only"
	@echo "  release       - Build optimized release binaries"
	@echo "  downloads     - Copy Linux binaries to downloads folder"
	@echo ""
	@echo "Tests:"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-agent    - Test agent module only"
	@echo "  test-pkg      - Test pkg modules only"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  security      - Run security scan"
	@echo "  vuln-check    - Check for vulnerabilities"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start services with Docker Compose"
	@echo "  docker-down   - Stop services"
	@echo "  docker-logs   - View service logs"
	@echo ""
	@echo "Development:"
	@echo "  dev-agent     - Run agent in development mode"
	@echo ""
	@echo "Utilities:"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  version       - Show current version"
	@echo "  help          - Show this help"
