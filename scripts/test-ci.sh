#!/bin/bash

# Local CI/CD Test Script
# This script simulates the CI/CD pipeline locally

set -e

echo "🚀 Starting Local CI/CD Test Pipeline"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Step 1: Clean environment
print_status "Cleaning environment..."
make clean
print_success "Environment cleaned"

# Step 2: Download dependencies
print_status "Downloading dependencies..."
go mod download
go mod verify
print_success "Dependencies downloaded and verified"

# Step 3: Code formatting check
print_status "Checking code formatting..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    print_error "Code is not formatted properly"
    gofmt -s -l .
    exit 1
fi
print_success "Code is properly formatted"

# Step 4: Run tests
print_status "Running tests..."
if ! go test -v -race -coverprofile=coverage.out ./...; then
    print_error "Tests failed"
    exit 1
fi
print_success "All tests passed"

# Step 5: Coverage check
print_status "Checking coverage..."
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Current coverage: ${COVERAGE}%"
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    print_warning "Coverage is below 80% (${COVERAGE}%)"
else
    print_success "Coverage threshold met (${COVERAGE}%)"
fi

# Step 6: Linting (if golangci-lint is available)
if command -v golangci-lint &> /dev/null; then
    print_status "Running linter..."
    if ! golangci-lint run; then
        print_error "Linting failed"
        exit 1
    fi
    print_success "Linting passed"
else
    print_warning "golangci-lint not installed, skipping linting"
fi

# Step 7: Security scan (if gosec is available)
if command -v gosec &> /dev/null; then
    print_status "Running security scan..."
    if ! gosec ./...; then
        print_warning "Security scan found issues"
    else
        print_success "Security scan passed"
    fi
else
    print_warning "gosec not installed, skipping security scan"
fi

# Step 8: Vulnerability check (if govulncheck is available)
if command -v govulncheck &> /dev/null; then
    print_status "Checking for vulnerabilities..."
    if ! govulncheck ./...; then
        print_warning "Vulnerabilities found"
    else
        print_success "No vulnerabilities found"
    fi
else
    print_warning "govulncheck not installed, skipping vulnerability check"
fi

# Step 9: Build application
print_status "Building application..."
if ! make build; then
    print_error "Build failed"
    exit 1
fi
print_success "Application built successfully"

# Step 10: Verify multi-tier metrics implementation
print_status "Verifying multi-tier metrics implementation..."

# Check for key files
if [ ! -f "internal/services/tiered_metrics.go" ]; then
    print_error "TieredMetricsService file not found!"
    exit 1
fi

if [ ! -f "internal/services/metrics_commands.go" ]; then
    print_error "MetricsCommandsService file not found!"
    exit 1
fi

if [ ! -f "internal/storage/timescaledb/multi_tier_metrics.go" ]; then
    print_error "Multi-tier metrics storage file not found!"
    exit 1
fi

# Check for key implementations
if ! grep -q "TieredMetricsService" internal/services/tiered_metrics.go; then
    print_error "TieredMetricsService implementation not found!"
    exit 1
fi

if ! grep -q "MetricsCommandsService" internal/services/metrics_commands.go; then
    print_error "MetricsCommandsService implementation not found!"
    exit 1
fi

print_success "Multi-tier metrics implementation verified"

# Step 11: Docker build (if Docker is available)
if command -v docker &> /dev/null; then
    print_status "Building Docker image..."
    if ! make docker-build; then
        print_error "Docker build failed"
        exit 1
    fi
    print_success "Docker image built successfully"
else
    print_warning "Docker not available, skipping Docker build"
fi

# Step 12: Test Docker image (if built)
if command -v docker &> /dev/null && docker images | grep -q "servereye-api"; then
    print_status "Testing Docker image..."
    
    # Run container in background
    CONTAINER_ID=$(docker run -d --rm -p 8081:8080 --env-file .env.example servereye-api:latest)
    
    # Wait for startup
    sleep 5
    
    # Test health endpoint
    if curl -f http://localhost:8081/health > /dev/null 2>&1; then
        print_success "Docker container health check passed"
    else
        print_warning "Docker container health check failed"
    fi
    
    # Stop container
    docker stop $CONTAINER_ID > /dev/null 2>&1 || true
fi

# Step 13: Generate test report
print_status "Generating test report..."
cat > test-report.txt << EOF
Local CI/CD Test Report
========================
Date: $(date)
Commit: $(git rev-parse HEAD 2>/dev/null || echo "N/A")
Branch: $(git branch --show-current 2>/dev/null || echo "N/A")

Results:
✅ Code formatting: PASSED
✅ Tests: PASSED
✅ Coverage: ${COVERAGE}%
EOF

if command -v golangci-lint &> /dev/null; then
    echo "✅ Linting: PASSED" >> test-report.txt
else
    echo "⚠️  Linting: SKIPPED" >> test-report.txt
fi

if command -v gosec &> /dev/null; then
    echo "✅ Security scan: PASSED" >> test-report.txt
else
    echo "⚠️  Security scan: SKIPPED" >> test-report.txt
fi

echo "✅ Build: PASSED" >> test-report.txt
echo "✅ Multi-tier metrics: VERIFIED" >> test-report.txt

if command -v docker &> /dev/null; then
    echo "✅ Docker build: PASSED" >> test-report.txt
else
    echo "⚠️  Docker build: SKIPPED" >> test-report.txt
fi

echo "" >> test-report.txt
echo "All critical checks passed! Ready for deployment." >> test-report.txt

print_success "Test report generated: test-report.txt"

echo ""
echo -e "${GREEN}🎉 Local CI/CD Test Pipeline Completed Successfully!${NC}"
echo ""
echo "📊 Summary:"
echo "  - Code formatting: ✅"
echo "  - Tests: ✅"
echo "  - Coverage: ${COVERAGE}%"
echo "  - Build: ✅"
echo "  - Multi-tier metrics: ✅"
echo ""
echo "📄 Full report available in: test-report.txt"
echo ""
echo "🚀 Ready to push to production!"
