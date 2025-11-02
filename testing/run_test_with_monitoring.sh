#!/bin/bash

# Combined Test and Monitoring Script
# Runs the Go test while simultaneously collecting CloudWatch metrics

echo "=========================================="
echo "MySQL Test with CloudWatch Monitoring"
echo "=========================================="
echo ""

# Check if test file exists
if [ ! -f "test.go" ]; then
    echo "Error: test.go not found"
    echo "Please ensure the test file is in the current directory"
    exit 1
fi

# Check if monitoring script exists
if [ ! -f "monitor_cloudwatch.sh" ]; then
    echo "Error: monitor_cloudwatch.sh not found"
    exit 1
fi

# Make scripts executable
chmod +x monitor_cloudwatch.sh

# Compile Go test
echo "Compiling test..."
go build test.go
if [ $? -ne 0 ]; then
    echo "Error: Failed to compile test"
    exit 1
fi

echo "✓ Test compiled successfully"
echo ""

# Start monitoring in background
echo "Starting CloudWatch monitoring..."
./monitor_cloudwatch.sh > monitoring.log 2>&1 &
MONITOR_PID=$!

# Wait a moment for monitoring to initialize
sleep 3

# Run the test
echo "Running test..."
echo ""
./test

# Wait for monitoring to complete
echo ""
echo "Waiting for monitoring to complete..."
wait $MONITOR_PID

# Display monitoring summary
echo ""
cat monitoring.log | grep -A 100 "CLOUDWATCH METRICS SUMMARY"

echo ""
echo "=========================================="
echo "Test and Monitoring Complete!"
echo "=========================================="
echo ""
echo "Generated files:"
echo "  - mysql_test_results.json (test results)"
echo "  - cloudwatch_metrics/ (all CloudWatch metrics)"
echo "  - monitoring.log (monitoring output)"
echo ""

# Generate comprehensive report
echo "Generating comprehensive report..."
python3 generate_report.py

if [ $? -eq 0 ]; then
    echo "✓ Comprehensive report generated: comprehensive_report.json"
else
    echo "⚠ Failed to generate comprehensive report"
    echo "You can run manually: python3 generate_report.py"
fi
echo ""