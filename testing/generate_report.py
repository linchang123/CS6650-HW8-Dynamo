#!/usr/bin/env python3
"""
Generate comprehensive test report combining:
- Test results (mysql_test_results.json)
- CloudWatch metrics (cloudwatch_metrics/)
"""

import json
import glob
import os
from datetime import datetime

def read_test_results():
    """Read test results JSON"""
    try:
        with open('test_results.json', 'r') as f:
            return json.load(f)
    except FileNotFoundError:
        print("Error: test_results.json not found")
        return None

def read_cloudwatch_metrics(pattern):
    """Read all CloudWatch metric files matching pattern"""
    files = glob.glob(os.path.join('cloudwatch_metrics', pattern))
    all_datapoints = []
    
    for file in sorted(files):
        try:
            with open(file, 'r') as f:
                data = json.load(f)
                if 'Datapoints' in data:
                    all_datapoints.extend(data['Datapoints'])
        except Exception as e:
            pass
    
    return all_datapoints

def calculate_stats(datapoints, stat_key='Average'):
    """Calculate statistics from CloudWatch datapoints"""
    if not datapoints:
        return None
    
    values = [dp[stat_key] for dp in datapoints if stat_key in dp]
    if not values:
        return None
    
    return {
        'min': round(min(values), 2),
        'max': round(max(values), 2),
        'avg': round(sum(values) / len(values), 2),
        'count': len(values)
    }

def generate_report():
    """Generate comprehensive report"""
    
    # Read test results
    test_results = read_test_results()
    if not test_results:
        return
    
    # Collect CloudWatch metrics
    metrics = {
        'rds': {
            'cpu': calculate_stats(read_cloudwatch_metrics('rds_cpu_*.json')),
            'connections': calculate_stats(read_cloudwatch_metrics('rds_connections_*.json')),
            'read_iops': calculate_stats(read_cloudwatch_metrics('rds_read_iops_*.json')),
            'write_iops': calculate_stats(read_cloudwatch_metrics('rds_write_iops_*.json')),
            'read_latency': calculate_stats(read_cloudwatch_metrics('rds_read_latency_*.json')),
            'write_latency': calculate_stats(read_cloudwatch_metrics('rds_write_latency_*.json')),
        },
        'ecs': {
            'cpu': calculate_stats(read_cloudwatch_metrics('ecs_cpu_*.json')),
            'memory': calculate_stats(read_cloudwatch_metrics('ecs_memory_*.json')),
        },
        'alb': {
            'response_time': calculate_stats(read_cloudwatch_metrics('alb_response_time_*.json')),
            'request_count': calculate_stats(read_cloudwatch_metrics('alb_request_count_*.json'), 'Sum'),
            'healthy_hosts': calculate_stats(read_cloudwatch_metrics('alb_healthy_hosts_*.json')),
        }
    }
    
    # Generate report
    report = {
        'report_generated': datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ'),
        'test_results': {
            'metadata': test_results.get('test_metadata', {}),
            'statistics': test_results.get('statistics', {})
        },
        'cloudwatch_metrics': metrics,
        'analysis': generate_analysis(test_results, metrics)
    }
    
    # Save report
    with open('comprehensive_report.json', 'w') as f:
        json.dump(report, f, indent=2)
    
    # Print summary
    print_report_summary(report)
    
    return report

def generate_analysis(test_results, metrics):
    """Generate performance analysis"""
    analysis = {
        'performance_grade': 'A',
        'issues': [],
        'recommendations': []
    }
    
    stats = test_results.get('statistics', {})
    
    # Check test success rate
    if stats.get('success_rate', 0) < 100:
        analysis['issues'].append(f"Test failures detected: {stats.get('failed_operations', 0)} operations failed")
        analysis['performance_grade'] = 'B'
    
    # Check RDS CPU
    rds_cpu = metrics.get('rds', {}).get('cpu')
    if rds_cpu and rds_cpu['avg'] > 70:
        analysis['issues'].append(f"High RDS CPU utilization: {rds_cpu['avg']}%")
        analysis['recommendations'].append("Consider upgrading RDS instance class")
        analysis['performance_grade'] = 'B'
    
    # Check RDS connections
    rds_conn = metrics.get('rds', {}).get('connections')
    if rds_conn and rds_conn['max'] > 50:
        analysis['issues'].append(f"High database connection count: {rds_conn['max']}")
        analysis['recommendations'].append("Review connection pooling settings")
    
    # Check ECS CPU
    ecs_cpu = metrics.get('ecs', {}).get('cpu')
    if ecs_cpu and ecs_cpu['avg'] > 70:
        analysis['issues'].append(f"High ECS CPU utilization: {ecs_cpu['avg']}%")
        analysis['recommendations'].append("Consider scaling ECS tasks or increasing CPU allocation")
        analysis['performance_grade'] = 'C'
    
    # Check response times
    ops = stats.get('operations', {})
    for op_name, op_stats in ops.items():
        if op_stats.get('avg_response_time', 0) > 1000:
            analysis['issues'].append(f"Slow {op_name} operations: {op_stats['avg_response_time']}ms avg")
            analysis['recommendations'].append(f"Investigate {op_name} query performance")
            if analysis['performance_grade'] == 'A':
                analysis['performance_grade'] = 'B'
    
    # Check database latency
    read_lat = metrics.get('rds', {}).get('read_latency')
    write_lat = metrics.get('rds', {}).get('write_latency')
    
    if read_lat and read_lat['avg'] > 10:
        analysis['issues'].append(f"High database read latency: {read_lat['avg']}ms")
        analysis['recommendations'].append("Consider adding database indexes or caching")
    
    if write_lat and write_lat['avg'] > 10:
        analysis['issues'].append(f"High database write latency: {write_lat['avg']}ms")
        analysis['recommendations'].append("Review database write operations and batch sizes")
    
    if not analysis['issues']:
        analysis['summary'] = "Excellent performance! All metrics within acceptable ranges."
    else:
        analysis['summary'] = f"Found {len(analysis['issues'])} performance issues."
    
    return analysis

def print_report_summary(report):
    """Print human-readable report summary"""
    
    print("\n" + "="*70)
    print("COMPREHENSIVE TEST REPORT")
    print("="*70)
    
    # Test Results
    test_stats = report['test_results']['statistics']
    print("\n📊 TEST RESULTS:")
    print("-"*70)
    print(f"Total Operations:     {test_stats.get('total_operations', 0)}")
    print(f"Successful:           {test_stats.get('successful_operations', 0)}")
    print(f"Failed:               {test_stats.get('failed_operations', 0)}")
    print(f"Success Rate:         {test_stats.get('success_rate', 0):.2f}%")
    
    # Response Times
    print("\n⏱️  RESPONSE TIMES:")
    print("-"*70)
    for op_name, op_stats in test_stats.get('operations', {}).items():
        print(f"{op_name:20s} avg: {op_stats.get('avg_response_time', 0):6.2f}ms  " +
              f"(min: {op_stats.get('min_response_time', 0):6.2f}ms, " +
              f"max: {op_stats.get('max_response_time', 0):6.2f}ms)")
    
    # RDS Metrics
    rds = report['cloudwatch_metrics']['rds']
    print("\n💾 RDS METRICS:")
    print("-"*70)
    if rds['cpu']:
        print(f"CPU Utilization:      {rds['cpu']['avg']:6.2f}% " +
              f"(min: {rds['cpu']['min']:6.2f}%, max: {rds['cpu']['max']:6.2f}%)")
    if rds['connections']:
        print(f"Database Connections: {rds['connections']['avg']:6.2f} " +
              f"(max: {rds['connections']['max']:6.2f})")
    if rds['read_iops']:
        print(f"Read IOPS:           {rds['read_iops']['avg']:6.2f} " +
              f"(max: {rds['read_iops']['max']:6.2f})")
    if rds['write_iops']:
        print(f"Write IOPS:          {rds['write_iops']['avg']:6.2f} " +
              f"(max: {rds['write_iops']['max']:6.2f})")
    if rds['read_latency']:
        print(f"Read Latency:        {rds['read_latency']['avg']:6.2f}ms " +
              f"(max: {rds['read_latency']['max']:6.2f}ms)")
    if rds['write_latency']:
        print(f"Write Latency:       {rds['write_latency']['avg']:6.2f}ms " +
              f"(max: {rds['write_latency']['max']:6.2f}ms)")
    
    # ECS Metrics
    ecs = report['cloudwatch_metrics']['ecs']
    print("\n🚀 ECS METRICS:")
    print("-"*70)
    if ecs['cpu']:
        print(f"CPU Utilization:     {ecs['cpu']['avg']:6.2f}% " +
              f"(max: {ecs['cpu']['max']:6.2f}%)")
    if ecs['memory']:
        print(f"Memory Utilization:  {ecs['memory']['avg']:6.2f}% " +
              f"(max: {ecs['memory']['max']:6.2f}%)")
    
    # ALB Metrics
    alb = report['cloudwatch_metrics']['alb']
    print("\n⚖️  ALB METRICS:")
    print("-"*70)
    if alb['response_time']:
        print(f"Response Time:       {alb['response_time']['avg']:6.2f}s " +
              f"(max: {alb['response_time']['max']:6.2f}s)")
    if alb['request_count']:
        print(f"Total Requests:      {int(alb['request_count']['max'])}")
    if alb['healthy_hosts']:
        print(f"Healthy Hosts (avg): {alb['healthy_hosts']['avg']:6.2f}")
    
    # Analysis
    analysis = report['analysis']
    print("\n📈 PERFORMANCE ANALYSIS:")
    print("-"*70)
    print(f"Grade: {analysis['performance_grade']}")
    print(f"\n{analysis['summary']}")
    
    if analysis['issues']:
        print("\n⚠️  Issues Found:")
        for i, issue in enumerate(analysis['issues'], 1):
            print(f"  {i}. {issue}")
    
    if analysis['recommendations']:
        print("\n💡 Recommendations:")
        for i, rec in enumerate(analysis['recommendations'], 1):
            print(f"  {i}. {rec}")
    
    print("\n" + "="*70)
    print("Report saved to: comprehensive_report.json")
    print("="*70)

if __name__ == "__main__":
    generate_report()