#!/bin/bash

# CloudWatch Metrics Monitoring Script
# Collects RDS, ECS, and ALB metrics during testing

# Configuration
REGION="us-west-2"
OUTPUT_DIR="cloudwatch_metrics"
INTERVAL=30  # Seconds between measurements
DURATION=300 # Total monitoring duration (5 minutes)

# Resource identifiers (hardcoded)
RDS_IDENTIFIER="cs6650l2-db"
ECS_CLUSTER="cs6650l2-cluster"
ECS_SERVICE="cs6650l2"

echo "Monitoring resources:"
echo "  RDS: $RDS_IDENTIFIER"
echo "  ECS Cluster: $ECS_CLUSTER"
echo "  ECS Service: $ECS_SERVICE"
echo "  Region: $REGION"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Get ALB ARN suffix
ALB_ARN_SUFFIX=$(aws elbv2 describe-load-balancers \
    --region "$REGION" \
    --query "LoadBalancers[?contains(LoadBalancerName, 'cs6650l2')].LoadBalancerArn" \
    --output text 2>/dev/null | awk -F: '{print $NF}')

if [ -n "$ALB_ARN_SUFFIX" ]; then
    echo "  ALB ARN Suffix: $ALB_ARN_SUFFIX"
else
    echo "  Warning: Could not find ALB ARN suffix"
fi

echo ""
echo "Starting monitoring for $DURATION seconds (interval: ${INTERVAL}s)..."
echo "Metrics will be saved to: $OUTPUT_DIR/"
echo ""

# Calculate end time
END_TIME=$(($(date +%s) + DURATION))
ITERATION=0

# Monitoring loop
while [ $(date +%s) -lt $END_TIME ]; do
    ITERATION=$((ITERATION + 1))
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    echo "[$TIMESTAMP] Collecting metrics (iteration $ITERATION)..."
    
    # Get current time for CloudWatch queries (last 5 minutes)
    END_QUERY_TIME=$(date -u +"%Y-%m-%dT%H:%M:%S")
    START_QUERY_TIME=$(date -u -v-5M +"%Y-%m-%dT%H:%M:%S" 2>/dev/null || date -u -d '5 minutes ago' +"%Y-%m-%dT%H:%M:%S")
    
    # 1. RDS Metrics
    echo "  - Collecting RDS metrics..."
    
    # RDS CPU Utilization
    aws cloudwatch get-metric-statistics \
        --namespace AWS/RDS \
        --metric-name CPUUtilization \
        --dimensions Name=DBInstanceIdentifier,Value="$RDS_IDENTIFIER" \
        --start-time "$START_QUERY_TIME" \
        --end-time "$END_QUERY_TIME" \
        --period 60 \
        --statistics Average Maximum \
        --region "$REGION" \
        > "$OUTPUT_DIR/rds_cpu_${ITERATION}.json" 2>/dev/null
    
    # RDS Database Connections
    aws cloudwatch get-metric-statistics \
        --namespace AWS/RDS \
        --metric-name DatabaseConnections \
        --dimensions Name=DBInstanceIdentifier,Value="$RDS_IDENTIFIER" \
        --start-time "$START_QUERY_TIME" \
        --end-time "$END_QUERY_TIME" \
        --period 60 \
        --statistics Average Maximum \
        --region "$REGION" \
        > "$OUTPUT_DIR/rds_connections_${ITERATION}.json" 2>/dev/null
    
    # RDS Read/Write IOPS
    aws cloudwatch get-metric-statistics \
        --namespace AWS/RDS \
        --metric-name ReadIOPS \
        --dimensions Name=DBInstanceIdentifier,Value="$RDS_IDENTIFIER" \
        --start-time "$START_QUERY_TIME" \
        --end-time "$END_QUERY_TIME" \
        --period 60 \
        --statistics Average Maximum \
        --region "$REGION" \
        > "$OUTPUT_DIR/rds_read_iops_${ITERATION}.json" 2>/dev/null
    
    aws cloudwatch get-metric-statistics \
        --namespace AWS/RDS \
        --metric-name WriteIOPS \
        --dimensions Name=DBInstanceIdentifier,Value="$RDS_IDENTIFIER" \
        --start-time "$START_QUERY_TIME" \
        --end-time "$END_QUERY_TIME" \
        --period 60 \
        --statistics Average Maximum \
        --region "$REGION" \
        > "$OUTPUT_DIR/rds_write_iops_${ITERATION}.json" 2>/dev/null
    
    # RDS Read/Write Latency
    aws cloudwatch get-metric-statistics \
        --namespace AWS/RDS \
        --metric-name ReadLatency \
        --dimensions Name=DBInstanceIdentifier,Value="$RDS_IDENTIFIER" \
        --start-time "$START_QUERY_TIME" \
        --end-time "$END_QUERY_TIME" \
        --period 60 \
        --statistics Average Maximum \
        --region "$REGION" \
        > "$OUTPUT_DIR/rds_read_latency_${ITERATION}.json" 2>/dev/null
    
    aws cloudwatch get-metric-statistics \
        --namespace AWS/RDS \
        --metric-name WriteLatency \
        --dimensions Name=DBInstanceIdentifier,Value="$RDS_IDENTIFIER" \
        --start-time "$START_QUERY_TIME" \
        --end-time "$END_QUERY_TIME" \
        --period 60 \
        --statistics Average Maximum \
        --region "$REGION" \
        > "$OUTPUT_DIR/rds_write_latency_${ITERATION}.json" 2>/dev/null
    
    # 2. ECS Metrics
    echo "  - Collecting ECS metrics..."
    
    # ECS CPU Utilization
    aws cloudwatch get-metric-statistics \
        --namespace AWS/ECS \
        --metric-name CPUUtilization \
        --dimensions Name=ServiceName,Value="$ECS_SERVICE" Name=ClusterName,Value="$ECS_CLUSTER" \
        --start-time "$START_QUERY_TIME" \
        --end-time "$END_QUERY_TIME" \
        --period 60 \
        --statistics Average Maximum \
        --region "$REGION" \
        > "$OUTPUT_DIR/ecs_cpu_${ITERATION}.json" 2>/dev/null
    
    # ECS Memory Utilization
    aws cloudwatch get-metric-statistics \
        --namespace AWS/ECS \
        --metric-name MemoryUtilization \
        --dimensions Name=ServiceName,Value="$ECS_SERVICE" Name=ClusterName,Value="$ECS_CLUSTER" \
        --start-time "$START_QUERY_TIME" \
        --end-time "$END_QUERY_TIME" \
        --period 60 \
        --statistics Average Maximum \
        --region "$REGION" \
        > "$OUTPUT_DIR/ecs_memory_${ITERATION}.json" 2>/dev/null
    
    # 3. ALB Metrics
    if [ -n "$ALB_ARN_SUFFIX" ]; then
        echo "  - Collecting ALB metrics..."
        
        # ALB Target Response Time
        aws cloudwatch get-metric-statistics \
            --namespace AWS/ApplicationELB \
            --metric-name TargetResponseTime \
            --dimensions Name=LoadBalancer,Value="$ALB_ARN_SUFFIX" \
            --start-time "$START_QUERY_TIME" \
            --end-time "$END_QUERY_TIME" \
            --period 60 \
            --statistics Average Maximum \
            --region "$REGION" \
            > "$OUTPUT_DIR/alb_response_time_${ITERATION}.json" 2>/dev/null
        
        # ALB Request Count
        aws cloudwatch get-metric-statistics \
            --namespace AWS/ApplicationELB \
            --metric-name RequestCount \
            --dimensions Name=LoadBalancer,Value="$ALB_ARN_SUFFIX" \
            --start-time "$START_QUERY_TIME" \
            --end-time "$END_QUERY_TIME" \
            --period 60 \
            --statistics Sum \
            --region "$REGION" \
            > "$OUTPUT_DIR/alb_request_count_${ITERATION}.json" 2>/dev/null
        
        # ALB Healthy/Unhealthy Host Count
        aws cloudwatch get-metric-statistics \
            --namespace AWS/ApplicationELB \
            --metric-name HealthyHostCount \
            --dimensions Name=LoadBalancer,Value="$ALB_ARN_SUFFIX" \
            --start-time "$START_QUERY_TIME" \
            --end-time "$END_QUERY_TIME" \
            --period 60 \
            --statistics Average \
            --region "$REGION" \
            > "$OUTPUT_DIR/alb_healthy_hosts_${ITERATION}.json" 2>/dev/null
    fi
    
    # Quick summary to console
    echo "  ✓ Metrics collected"
    
    # Sleep until next interval (unless this is the last iteration)
    if [ $(date +%s) -lt $END_TIME ]; then
        sleep "$INTERVAL"
    fi
done

echo ""
echo "Monitoring complete!"
echo ""
echo "Generating summary report..."

# Generate summary report
python3 - <<EOF
import json
import glob
import os

output_dir = "$OUTPUT_DIR"

def read_metric_files(pattern):
    """Read all JSON files matching pattern and extract datapoints"""
    files = glob.glob(os.path.join(output_dir, pattern))
    all_datapoints = []
    
    for file in files:
        try:
            with open(file, 'r') as f:
                data = json.load(f)
                if 'Datapoints' in data:
                    all_datapoints.extend(data['Datapoints'])
        except:
            pass
    
    return all_datapoints

def calculate_stats(datapoints, stat_key='Average'):
    """Calculate min, max, avg from datapoints"""
    if not datapoints:
        return None
    
    values = [dp[stat_key] for dp in datapoints if stat_key in dp]
    if not values:
        return None
    
    return {
        'min': round(min(values), 2),
        'max': round(max(values), 2),
        'avg': round(sum(values) / len(values), 2),
        'samples': len(values)
    }

print("\n" + "="*60)
print("CLOUDWATCH METRICS SUMMARY")
print("="*60)

# RDS Metrics
print("\n📊 RDS Metrics:")
print("-" * 60)

rds_cpu = read_metric_files('rds_cpu_*.json')
cpu_stats = calculate_stats(rds_cpu)
if cpu_stats:
    print(f"CPU Utilization:      {cpu_stats['avg']}% (min: {cpu_stats['min']}%, max: {cpu_stats['max']}%)")

rds_conn = read_metric_files('rds_connections_*.json')
conn_stats = calculate_stats(rds_conn)
if conn_stats:
    print(f"Database Connections: {conn_stats['avg']} (min: {conn_stats['min']}, max: {conn_stats['max']})")

rds_read_iops = read_metric_files('rds_read_iops_*.json')
read_iops_stats = calculate_stats(rds_read_iops)
if read_iops_stats:
    print(f"Read IOPS:           {read_iops_stats['avg']} (max: {read_iops_stats['max']})")

rds_write_iops = read_metric_files('rds_write_iops_*.json')
write_iops_stats = calculate_stats(rds_write_iops)
if write_iops_stats:
    print(f"Write IOPS:          {write_iops_stats['avg']} (max: {write_iops_stats['max']})")

rds_read_lat = read_metric_files('rds_read_latency_*.json')
read_lat_stats = calculate_stats(rds_read_lat)
if read_lat_stats:
    print(f"Read Latency:        {read_lat_stats['avg']}ms (max: {read_lat_stats['max']}ms)")

rds_write_lat = read_metric_files('rds_write_latency_*.json')
write_lat_stats = calculate_stats(rds_write_lat)
if write_lat_stats:
    print(f"Write Latency:       {write_lat_stats['avg']}ms (max: {write_lat_stats['max']}ms)")

# ECS Metrics
print("\n📊 ECS Metrics:")
print("-" * 60)

ecs_cpu = read_metric_files('ecs_cpu_*.json')
ecs_cpu_stats = calculate_stats(ecs_cpu)
if ecs_cpu_stats:
    print(f"CPU Utilization:     {ecs_cpu_stats['avg']}% (min: {ecs_cpu_stats['min']}%, max: {ecs_cpu_stats['max']}%)")

ecs_mem = read_metric_files('ecs_memory_*.json')
ecs_mem_stats = calculate_stats(ecs_mem)
if ecs_mem_stats:
    print(f"Memory Utilization:  {ecs_mem_stats['avg']}% (min: {ecs_mem_stats['min']}%, max: {ecs_mem_stats['max']}%)")

# ALB Metrics
print("\n📊 ALB Metrics:")
print("-" * 60)

alb_response = read_metric_files('alb_response_time_*.json')
alb_response_stats = calculate_stats(alb_response)
if alb_response_stats:
    print(f"Response Time:       {alb_response_stats['avg']}s (max: {alb_response_stats['max']}s)")

alb_requests = read_metric_files('alb_request_count_*.json')
alb_req_stats = calculate_stats(alb_requests, 'Sum')
if alb_req_stats:
    print(f"Total Requests:      {int(alb_req_stats['max'])}")

alb_healthy = read_metric_files('alb_healthy_hosts_*.json')
alb_healthy_stats = calculate_stats(alb_healthy)
if alb_healthy_stats:
    print(f"Healthy Hosts:       {alb_healthy_stats['avg']} (avg)")

print("\n" + "="*60)
print(f"All metrics saved to: {output_dir}/")
print("="*60)
EOF

echo ""
echo "✓ Monitoring complete!"
echo "✓ Metrics saved to: $OUTPUT_DIR/"