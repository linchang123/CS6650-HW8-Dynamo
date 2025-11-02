#!/bin/bash

# DynamoDB CloudWatch Monitoring Script
# Collects metrics during test execution

echo "=========================================="
echo "DynamoDB CloudWatch Monitoring"
echo "=========================================="
echo ""

# Configuration
REGION="us-west-2"
PRODUCTS_TABLE="ecommerce-products"
CARTS_TABLE="ecommerce-carts"
ECS_CLUSTER="cs6650l2-cluster"
ECS_SERVICE="cs6650l2"
ALB_ARN_SUFFIX="app/cs6650l2-alb/XXXXX"  # Will be auto-detected
INTERVAL=30  # Collect metrics every 30 seconds
DURATION=300 # Monitor for 5 minutes (same as test limit)
OUTPUT_DIR="cloudwatch_metrics_dynamodb"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "Configuration:"
echo "  Region: $REGION"
echo "  Products Table: $PRODUCTS_TABLE"
echo "  Carts Table: $CARTS_TABLE"
echo "  ECS Cluster: $ECS_CLUSTER"
echo "  ECS Service: $ECS_SERVICE"
echo "  Interval: ${INTERVAL}s"
echo "  Duration: ${DURATION}s"
echo "  Output: $OUTPUT_DIR/"
echo ""

# Auto-detect ALB ARN
echo "Detecting ALB ARN..."
ALB_FULL_ARN=$(aws elbv2 describe-load-balancers \
    --region $REGION \
    --query "LoadBalancers[?contains(LoadBalancerName, 'cs6650l2-alb')].LoadBalancerArn" \
    --output text 2>/dev/null)

if [ -n "$ALB_FULL_ARN" ]; then
    # Extract the suffix (everything after loadbalancer/)
    ALB_ARN_SUFFIX=$(echo $ALB_FULL_ARN | sed 's/.*loadbalancer\///')
    echo "  ALB ARN Suffix: $ALB_ARN_SUFFIX"
else
    echo "  Warning: Could not detect ALB ARN, ALB metrics will be skipped"
fi
echo ""

# Start and end times
START_TIME=$(date -u +%Y-%m-%dT%H:%M:%S)
END_EPOCH=$(($(date +%s) + DURATION))

echo "Started monitoring at: $START_TIME"
echo "Will monitor until: $(date -r $END_EPOCH -u +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%S)"
echo ""

# Counter for collection rounds
ROUND=1

# Monitor until duration expires
while [ $(date +%s) -lt $END_EPOCH ]; do
    CURRENT_TIME=$(date -u +%Y-%m-%dT%H:%M:%S)
    echo "[$CURRENT_TIME] Collecting metrics (Round $ROUND)..."
    
    # Calculate time window (last 60 seconds)
    END_TIME=$(date -u +%Y-%m-%dT%H:%M:%S)
    # macOS compatible: use -v for relative time
    START_TIME=$(date -u -v-60S +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u -d '60 seconds ago' +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%S)
    
    # DynamoDB Metrics for Products Table
    # ConsumedReadCapacityUnits
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name ConsumedReadCapacityUnits \
        --dimensions Name=TableName,Value=$PRODUCTS_TABLE \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Sum \
        --region $REGION \
        > "$OUTPUT_DIR/products_read_capacity_${ROUND}.json" 2>/dev/null
    
    # ConsumedWriteCapacityUnits
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name ConsumedWriteCapacityUnits \
        --dimensions Name=TableName,Value=$PRODUCTS_TABLE \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Sum \
        --region $REGION \
        > "$OUTPUT_DIR/products_write_capacity_${ROUND}.json" 2>/dev/null
    
    # SuccessfulRequestLatency for GetItem
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name SuccessfulRequestLatency \
        --dimensions Name=TableName,Value=$PRODUCTS_TABLE Name=Operation,Value=GetItem \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Average Maximum \
        --region $REGION \
        > "$OUTPUT_DIR/products_getitem_latency_${ROUND}.json" 2>/dev/null
    
    # DynamoDB Metrics for Carts Table
    # ConsumedReadCapacityUnits
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name ConsumedReadCapacityUnits \
        --dimensions Name=TableName,Value=$CARTS_TABLE \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Sum \
        --region $REGION \
        > "$OUTPUT_DIR/carts_read_capacity_${ROUND}.json" 2>/dev/null
    
    # ConsumedWriteCapacityUnits
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name ConsumedWriteCapacityUnits \
        --dimensions Name=TableName,Value=$CARTS_TABLE \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Sum \
        --region $REGION \
        > "$OUTPUT_DIR/carts_write_capacity_${ROUND}.json" 2>/dev/null
    
    # SuccessfulRequestLatency for PutItem
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name SuccessfulRequestLatency \
        --dimensions Name=TableName,Value=$CARTS_TABLE Name=Operation,Value=PutItem \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Average Maximum \
        --region $REGION \
        > "$OUTPUT_DIR/carts_putitem_latency_${ROUND}.json" 2>/dev/null
    
    # SuccessfulRequestLatency for GetItem
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name SuccessfulRequestLatency \
        --dimensions Name=TableName,Value=$CARTS_TABLE Name=Operation,Value=GetItem \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Average Maximum \
        --region $REGION \
        > "$OUTPUT_DIR/carts_getitem_latency_${ROUND}.json" 2>/dev/null
    
    # UserErrors (throttling)
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name UserErrors \
        --dimensions Name=TableName,Value=$CARTS_TABLE \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Sum \
        --region $REGION \
        > "$OUTPUT_DIR/carts_user_errors_${ROUND}.json" 2>/dev/null
    
    aws cloudwatch get-metric-statistics \
        --namespace AWS/DynamoDB \
        --metric-name UserErrors \
        --dimensions Name=TableName,Value=$PRODUCTS_TABLE \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Sum \
        --region $REGION \
        > "$OUTPUT_DIR/products_user_errors_${ROUND}.json" 2>/dev/null
    
    # ECS Metrics
    # CPU Utilization
    aws cloudwatch get-metric-statistics \
        --namespace AWS/ECS \
        --metric-name CPUUtilization \
        --dimensions Name=ServiceName,Value=$ECS_SERVICE Name=ClusterName,Value=$ECS_CLUSTER \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Average Maximum \
        --region $REGION \
        > "$OUTPUT_DIR/ecs_cpu_${ROUND}.json" 2>/dev/null
    
    # Memory Utilization
    aws cloudwatch get-metric-statistics \
        --namespace AWS/ECS \
        --metric-name MemoryUtilization \
        --dimensions Name=ServiceName,Value=$ECS_SERVICE Name=ClusterName,Value=$ECS_CLUSTER \
        --start-time $START_TIME \
        --end-time $END_TIME \
        --period 60 \
        --statistics Average Maximum \
        --region $REGION \
        > "$OUTPUT_DIR/ecs_memory_${ROUND}.json" 2>/dev/null
    
    # ALB Metrics (only if ALB ARN was detected)
    if [ -n "$ALB_ARN_SUFFIX" ]; then
        # Target Response Time
        aws cloudwatch get-metric-statistics \
            --namespace AWS/ApplicationELB \
            --metric-name TargetResponseTime \
            --dimensions Name=LoadBalancer,Value=$ALB_ARN_SUFFIX \
            --start-time $START_TIME \
            --end-time $END_TIME \
            --period 60 \
            --statistics Average Maximum \
            --region $REGION \
            > "$OUTPUT_DIR/alb_response_time_${ROUND}.json" 2>/dev/null
        
        # Request Count
        aws cloudwatch get-metric-statistics \
            --namespace AWS/ApplicationELB \
            --metric-name RequestCount \
            --dimensions Name=LoadBalancer,Value=$ALB_ARN_SUFFIX \
            --start-time $START_TIME \
            --end-time $END_TIME \
            --period 60 \
            --statistics Sum \
            --region $REGION \
            > "$OUTPUT_DIR/alb_request_count_${ROUND}.json" 2>/dev/null
        
        # Healthy Host Count
        aws cloudwatch get-metric-statistics \
            --namespace AWS/ApplicationELB \
            --metric-name HealthyHostCount \
            --dimensions Name=LoadBalancer,Value=$ALB_ARN_SUFFIX Name=TargetGroup,Value=targetgroup/cs6650l2-tg/* \
            --start-time $START_TIME \
            --end-time $END_TIME \
            --period 60 \
            --statistics Average \
            --region $REGION \
            > "$OUTPUT_DIR/alb_healthy_hosts_${ROUND}.json" 2>/dev/null
    fi
    
    echo "  ✓ Round $ROUND complete"
    
    ROUND=$((ROUND + 1))
    
    # Sleep until next collection (unless we're done)
    if [ $(date +%s) -lt $END_EPOCH ]; then
        sleep $INTERVAL
    fi
done

echo ""
echo "=========================================="
echo "Monitoring Complete!"
echo "=========================================="
echo ""
echo "Collected metrics in: $OUTPUT_DIR/"
echo "Total collection rounds: $((ROUND - 1))"
echo ""

# Generate summary
echo "CLOUDWATCH METRICS SUMMARY"
echo "=========================================="
echo ""

echo "Products Table Metrics:"
echo "  Read Capacity: $(ls $OUTPUT_DIR/products_read_capacity_*.json 2>/dev/null | wc -l) data points"
echo "  Write Capacity: $(ls $OUTPUT_DIR/products_write_capacity_*.json 2>/dev/null | wc -l) data points"
echo "  GetItem Latency: $(ls $OUTPUT_DIR/products_getitem_latency_*.json 2>/dev/null | wc -l) data points"
echo ""

echo "Carts Table Metrics:"
echo "  Read Capacity: $(ls $OUTPUT_DIR/carts_read_capacity_*.json 2>/dev/null | wc -l) data points"
echo "  Write Capacity: $(ls $OUTPUT_DIR/carts_write_capacity_*.json 2>/dev/null | wc -l) data points"
echo "  PutItem Latency: $(ls $OUTPUT_DIR/carts_putitem_latency_*.json 2>/dev/null | wc -l) data points"
echo "  GetItem Latency: $(ls $OUTPUT_DIR/carts_getitem_latency_*.json 2>/dev/null | wc -l) data points"
echo "  User Errors: $(ls $OUTPUT_DIR/carts_user_errors_*.json 2>/dev/null | wc -l) data points"
echo ""

echo "ECS Metrics:"
echo "  CPU Utilization: $(ls $OUTPUT_DIR/ecs_cpu_*.json 2>/dev/null | wc -l) data points"
echo "  Memory Utilization: $(ls $OUTPUT_DIR/ecs_memory_*.json 2>/dev/null | wc -l) data points"
echo ""

if [ -n "$ALB_ARN_SUFFIX" ]; then
    echo "ALB Metrics:"
    echo "  Response Time: $(ls $OUTPUT_DIR/alb_response_time_*.json 2>/dev/null | wc -l) data points"
    echo "  Request Count: $(ls $OUTPUT_DIR/alb_request_count_*.json 2>/dev/null | wc -l) data points"
    echo "  Healthy Hosts: $(ls $OUTPUT_DIR/alb_healthy_hosts_*.json 2>/dev/null | wc -l) data points"
    echo ""
fi

echo "To analyze metrics, run:"
echo "  python3 generate_dynamodb_report.py"
echo ""