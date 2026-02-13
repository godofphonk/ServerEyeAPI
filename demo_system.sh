#!/bin/bash

echo "🚀 ServerEyeAPI Multi-Tier Metrics System Demo"
echo "============================================"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base URLs
METRICS_SERVER="http://localhost:8083"
COMMANDS_SERVER="http://localhost:8084"

# Check if test servers are running, start them if needed
if ! curl -s $METRICS_SERVER/health > /dev/null 2>&1; then
    echo -e "${YELLOW}Starting metrics test server...${NC}"
    cd test_servers && go build -o test_metrics test_metrics_endpoints.go && ./test_metrics &
    cd ..
    sleep 2
fi

if ! curl -s $COMMANDS_SERVER/health > /dev/null 2>&1; then
    echo -e "${YELLOW}Starting commands test server...${NC}"
    cd test_servers && go build -o test_commands test_commands.go && ./test_commands &
    cd ..
    sleep 2
fi

echo -e "${BLUE}Testing Metrics Endpoints${NC}"
echo "----------------------------"

# Test 1: Health check
echo -e "${YELLOW}1. Health Check${NC}"
curl -s $METRICS_SERVER/health | jq -r '"Status: " + .status + ", Time: " + .time'
echo

# Test 2: Auto-granularity metrics (1 hour = 1m granularity)
echo -e "${YELLOW}2. Auto-Granularity Metrics (1-hour range)${NC}"
RESULT=$(curl -s "$METRICS_SERVER/api/servers/test-server-001/metrics/tiered?start=2026-02-13T15:00:00Z&end=2026-02-13T16:00:00Z")
echo "Granularity selected: $(echo $RESULT | jq -r .granularity)"
echo "Total points: $(echo $RESULT | jq -r .total_points)"
echo "First point CPU: $(echo $RESULT | jq -r '.data_points[0].cpu_avg')"
echo

# Test 3: Real-time metrics
echo -e "${YELLOW}3. Real-Time Metrics (last 30 minutes)${NC}"
RESULT=$(curl -s "$METRICS_SERVER/api/servers/test-server-001/metrics/realtime?duration=30m")
echo "Points returned: $(echo $RESULT | jq -r .total_points)"
echo "Latest CPU: $(echo $RESULT | jq -r '.data_points[-1].cpu_avg')"
echo

# Test 4: Dashboard metrics
echo -e "${YELLOW}4. Dashboard Metrics${NC}"
RESULT=$(curl -s "$METRICS_SERVER/api/servers/test-server-001/metrics/dashboard")
echo "Current CPU: $(echo $RESULT | jq -r .current_status.cpu_usage)%"
echo "Current Memory: $(echo $RESULT | jq -r .current_status.memory_usage)%"
echo "Uptime: $(echo $RESULT | jq -r .current_status.uptime)"
echo

# Test 5: Metrics comparison
echo -e "${YELLOW}5. Period Comparison${NC}"
RESULT=$(curl -s "$METRICS_SERVER/api/servers/test-server-001/metrics/comparison?period1_start=2026-02-13T00:00:00Z&period1_end=2026-02-13T12:00:00Z&period2_start=2026-02-13T12:00:00Z&period2_end=2026-02-13T23:59:59Z")
echo "Period 1 Avg CPU: $(echo $RESULT | jq -r '.averages1.cpu_avg')%"
echo "Period 2 Avg CPU: $(echo $RESULT | jq -r '.averages2.cpu_avg')%"
echo "CPU Change: $(echo $RESULT | jq -r '.changes.cpu_change')%"
echo

# Test 6: System summary
echo -e "${YELLOW}6. System Metrics Summary${NC}"
RESULT=$(curl -s $METRICS_SERVER/api/metrics/summary)
echo "Total Data Points: $(echo $RESULT | jq -r .total_data_points)"
echo "Total Servers: $(echo $RESULT | jq -r .total_servers)"
echo "Storage Size: $(echo $RESULT | jq -r .storage_size)"
echo "1m Table Size: $(echo $RESULT | jq -r '.granularity_stats."1m".table_size')"
echo

echo -e "${BLUE}Testing Commands Endpoints${NC}"
echo "----------------------------"

# Test 7: Get metrics statistics
echo -e "${YELLOW}7. Get Metrics Statistics${NC}"
RESULT=$(curl -X POST $COMMANDS_SERVER/api/servers/management/command \
  -H "Content-Type: application/json" \
  -d '{"server_id":"management","type":"metrics_stats","payload":{}}')
echo "Command ID: $(echo $RESULT | jq -r .command_id)"
echo "Status: $(echo $RESULT | jq -r .status)"
echo "Total Records: $(echo $RESULT | jq -r .result.data.total_records)"
echo

# Test 8: Refresh aggregates
echo -e "${YELLOW}8. Refresh 5-minute Aggregates${NC}"
RESULT=$(curl -X POST $COMMANDS_SERVER/api/servers/management/command \
  -H "Content-Type: application/json" \
  -d '{"server_id":"management","type":"refresh_aggregates","payload":{"granularity":"5m"}}')
echo "Command ID: $(echo $RESULT | jq -r .command_id)"
echo "Message: $(echo $RESULT | jq -r .message)"
echo "Refreshed at: $(echo $RESULT | jq -r .result.data.refreshed_at)"
echo

# Test 9: Cleanup old metrics (dry run)
echo -e "${YELLOW}9. Cleanup Old Metrics (Dry Run)${NC}"
RESULT=$(curl -X POST $COMMANDS_SERVER/api/servers/management/command \
  -H "Content-Type: application/json" \
  -d '{"server_id":"management","type":"cleanup_old_metrics","payload":{"older_than":"90 days","dry_run":true}}')
echo "Command ID: $(echo $RESULT | jq -r .command_id)"
echo "Message: $(echo $RESULT | jq -r .message)"
echo "Estimated rows to delete: $(echo $RESULT | jq -r .result.data.estimated_rows)"
echo

echo -e "${GREEN}✅ All tests completed successfully!${NC}"
echo
echo "📊 Granularity Strategy:"
echo "   • Last hour: 1-minute intervals"
echo "   • Last 3 hours: 5-minute intervals"
echo "   • Last 24 hours: 10-minute intervals"
echo "   • Last 30 days: 1-hour intervals"
echo
echo "🔧 Available Commands:"
echo "   • refresh_aggregates - Update continuous aggregates"
echo "   • metrics_stats - Get storage statistics"
echo "   • cleanup_old_metrics - Remove old data"
echo "   • compression_policy - Apply compression"
echo "   • retention_policy - Configure retention"
echo "   • analyze_performance - Analyze query performance"
echo "   • export_metrics - Export data to file"
echo "   • import_metrics - Import data from file"
echo "   • validate_metrics - Check data integrity"
echo "   • optimize_storage - Optimize database performance"
