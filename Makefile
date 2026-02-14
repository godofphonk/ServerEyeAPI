# ServerEye API Makefile

.PHONY: build build-api test test-coverage test-coverage-threshold test-integration test-all install-coverage-tools clean docker-build docker-run docker-stop lint security vuln-check fmt

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
API_BINARY=servereye-api

# Build directories
BUILD_DIR=build

# Default target
all: build

# Version and build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS = -X github.com/godofphonk/ServerEyeAPI/internal/version.Version=$(VERSION) \
	-X github.com/godofphonk/ServerEyeAPI/internal/version.BuildTime=$(BUILD_DATE) \
	-X github.com/godofphonk/ServerEyeAPI/internal/version.GitCommit=$(GIT_COMMIT)

# Build API only
build: build-api

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
	go generate ./internal/wire
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(API_BINARY) ./cmd/api
	./$(BUILD_DIR)/$(API_BINARY)

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

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf coverage

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Security scan (requires gosec)
security:
	@echo "Running security scan..."
	gosec ./...

# Check for vulnerabilities (requires govulncheck)
vuln-check:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

# Generate mocks and Wire dependencies
mocks:
	@echo "Generating mocks and Wire dependencies..."
	go generate ./...

# Docker targets
docker-build:
	docker build -t servereye-api:latest .

docker-run:
	docker run -p 8080:8080 --env-file .env servereye-api:latest

docker-push:
	docker tag servereye-api:latest ghcr.io/godofphonk/servereyeapi:latest
	docker push ghcr.io/godofphonk/servereyeapi:latest

# Docker Compose targets
docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-compose-logs:
	docker-compose logs -f

# Release build with optimizations and version
RELEASE_LDFLAGS = -w -s \
	-X github.com/godofphonk/ServerEyeAPI/internal/version.Version=$(VERSION) \
	-X github.com/godofphonk/ServerEyeAPI/internal/version.BuildDate=$(BUILD_DATE) \
	-X github.com/godofphonk/ServerEyeAPI/internal/version.GitCommit=$(GIT_COMMIT)

release: clean
	@echo "Building release binary for version $(VERSION)..."
	@echo "Build date: $(BUILD_DATE)"
	@echo "Git commit: $(GIT_COMMIT)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(RELEASE_LDFLAGS)" -o $(BUILD_DIR)/$(API_BINARY)-linux-amd64 ./cmd/api
	@echo "✅ Release build complete!"
	@$(BUILD_DIR)/$(API_BINARY)-linux-amd64 --version 2>/dev/null || echo "Binary built successfully"

# Show current version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Git Commit: $(GIT_COMMIT)"

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build:"
	@echo "  build         - Build API"
	@echo "  build-api     - Build API only"
	@echo "  release       - Build optimized release binary"
	@echo "  run           - Run API locally"
	@echo ""
	@echo "Tests:"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all      - Run tests with coverage"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  security      - Run security scan"
	@echo "  vuln-check    - Check for vulnerabilities"
	@echo "  mocks         - Generate mocks and dependencies"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-push   - Push to registry"
	@echo "  docker-compose-up - Start services with Docker Compose"
	@echo "  docker-compose-down - Stop services"
	@echo "  docker-compose-logs - View service logs"
	@echo ""
	@echo "Utilities:"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  version       - Show current version"
	@echo "  help          - Show this help"
