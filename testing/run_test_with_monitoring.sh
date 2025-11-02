#!/bin/bash

# Combined DynamoDB Test and Monitoring Script
# Runs the Go test while simultaneously collecting CloudWatch metrics

echo "=========================================="
echo "DynamoDB Test with CloudWatch Monitoring"
echo "=========================================="
echo ""

# Get ALB URL from Terraform output
ALB_URL=$(cd ../terraform && terraform output -raw application_url 2>/dev/null)

if [ -z "$ALB_URL" ]; then
    echo "Error: Could not get ALB URL from Terraform"
    echo "Please provide ALB URL as first argument:"
    echo "  ./run_test_with_monitoring.sh http://your-alb-url.amazonaws.com"
    
    if [ -n "$1" ]; then
        ALB_URL="$1"
        echo ""
        echo "Using provided URL: $ALB_URL"
    else
        exit 1
    fi
fi

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

echo "Configuration:"
echo "  Target URL: $ALB_URL"
echo "  Test: test.go"
echo "  Monitoring: monitor_cloudwatch.sh"
echo ""

# Start monitoring in background
echo "Starting CloudWatch monitoring..."
./monitor_cloudwatch.sh > monitoring_dynamodb.log 2>&1 &
MONITOR_PID=$!

# Wait a moment for monitoring to initialize
sleep 3

# Run the test
echo "Running DynamoDB test..."
echo ""
go run test.go "$ALB_URL"
TEST_EXIT_CODE=$?

# Wait for monitoring to complete
echo ""
echo "Waiting for monitoring to complete..."
wait $MONITOR_PID

# Display monitoring summary
echo ""
cat monitoring_dynamodb.log | grep -A 100 "CLOUDWATCH METRICS SUMMARY"

echo ""
echo "=========================================="
echo "Test and Monitoring Complete!"
echo "=========================================="
echo ""
echo "Generated files:"
echo "  - dynamodb_test_results.json (test results)"
echo "  - cloudwatch_metrics_dynamodb/ (all CloudWatch metrics)"
echo "  - monitoring_dynamodb.log (monitoring output)"
echo ""

# Generate comprehensive report
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo "Generating comprehensive report..."
    python3 generate_report.py
    
    if [ $? -eq 0 ]; then
        echo "✓ Comprehensive report generated: comprehensive_dynamodb_report.json"
    else
        echo "⚠ Failed to generate comprehensive report"
        echo "You can run manually: python3 generate_report.py"
    fi
else
    echo "⚠ Test failed, skipping report generation"
fi

echo ""
echo "To compare with MySQL results:"
echo "  # View DynamoDB performance"
echo "  cat comprehensive_dynamodb_report.json | jq '.performance_grade'"
echo "  "
echo "  # View MySQL performance"
echo "  cat comprehensive_report.json | jq '.performance_grade'"
echo "  "
echo "  # Compare response times"
echo "  echo 'DynamoDB:' && cat dynamodb_test_results.json | jq '.statistics'"
echo "  echo 'MySQL:' && cat mysql_test_results.json | jq '.statistics'"
echo ""