#!/bin/bash

# Test Coverage Script for ServerEye API
# This script runs tests with coverage and generates reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Coverage threshold
THRESHOLD=80

echo -e "${YELLOW}🧪 Running ServerEye API Test Coverage...${NC}"

# Create coverage directory
mkdir -p coverage

# Run tests with coverage
echo -e "${YELLOW}📊 Running tests with coverage...${NC}"
go test -v -race -coverprofile=coverage/coverage.out ./...

# Generate coverage report
echo -e "${YELLOW}📈 Generating coverage report...${NC}"
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Get coverage percentage
COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo -e "${YELLOW}📋 Current Coverage: ${COVERAGE}%${NC}"

# Check if coverage meets threshold
if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo -e "${RED}❌ Coverage is below ${THRESHOLD}%${NC}"
    echo -e "${RED}Please add more tests to improve coverage.${NC}"
    exit 1
else
    echo -e "${GREEN}✅ Coverage threshold met!${NC}"
fi

# Generate coverage report per package
echo -e "${YELLOW}📦 Generating per-package coverage...${NC}"
echo "" > coverage/coverage-by-package.txt

for pkg in $(go list ./...); do
    pkg_name=$(echo $pkg | sed 's|github.com/godofphonk/ServerEyeAPI/||')
    if [ "$pkg_name" != "" ]; then
        pkg_coverage=$(go test -coverprofile=coverage/tmp.out $pkg 2>/dev/null | grep -oP 'coverage: \K[0-9.]+' || echo "0.0")
        echo "$pkg_name: ${pkg_coverage}%" >> coverage/coverage-by-package.txt
    fi
done

# Display package coverage
echo -e "${YELLOW}📊 Coverage by Package:${NC}"
cat coverage/coverage-by-package.txt | column -t

# Clean up temporary files
rm -f coverage/tmp.out

echo -e "${GREEN}✅ Test coverage completed successfully!${NC}"
echo -e "${GREEN}📄 Coverage report: coverage/coverage.html${NC}"
echo -e "${GREEN}📊 Coverage by package: coverage/coverage-by-package.txt${NC}"
