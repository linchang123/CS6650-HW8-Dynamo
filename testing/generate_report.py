#!/usr/bin/env python3
"""
Generate comprehensive test report combining:
- Test results (dynamodb_test_results.json)
- CloudWatch metrics (cloudwatch_metrics_dynamodb/)
"""

import json
import glob
import os
from datetime import datetime

def read_test_results():
    """Read test results JSON"""
    try:
        with open('dynamodb_test_results.json', 'r') as f:
            return json.load(f)
    except FileNotFoundError:
        print("Error: dynamodb_test_results.json not found")
        return None

def read_cloudwatch_metrics(pattern):
    """Read all CloudWatch metric files matching pattern"""
    files = glob.glob(os.path.join('cloudwatch_metrics_dynamodb', pattern))
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
    
    # Collect CloudWatch metrics for DynamoDB
    metrics = {
        'products_table': {
            'read_capacity': calculate_stats(read_cloudwatch_metrics('products_read_capacity_*.json'), 'Sum'),
            'write_capacity': calculate_stats(read_cloudwatch_metrics('products_write_capacity_*.json'), 'Sum'),
            'getitem_latency': calculate_stats(read_cloudwatch_metrics('products_getitem_latency_*.json')),
            'user_errors': calculate_stats(read_cloudwatch_metrics('products_user_errors_*.json'), 'Sum'),
        },
        'carts_table': {
            'read_capacity': calculate_stats(read_cloudwatch_metrics('carts_read_capacity_*.json'), 'Sum'),
            'write_capacity': calculate_stats(read_cloudwatch_metrics('carts_write_capacity_*.json'), 'Sum'),
            'putitem_latency': calculate_stats(read_cloudwatch_metrics('carts_putitem_latency_*.json')),
            'getitem_latency': calculate_stats(read_cloudwatch_metrics('carts_getitem_latency_*.json')),
            'user_errors': calculate_stats(read_cloudwatch_metrics('carts_user_errors_*.json'), 'Sum'),
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
        'database_type': 'DynamoDB',
        'test_results': {
            'metadata': test_results.get('test_metadata', {}),
            'statistics': test_results.get('statistics', {})
        },
        'cloudwatch_metrics': metrics,
        'analysis': generate_analysis(test_results, metrics)
    }
    
    # Save report
    with open('comprehensive_dynamodb_report.json', 'w') as f:
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
    total_ops = len(test_results.get('results', []))
    success_rate = sum(1 for r in test_results.get('results', []) if r['success']) / total_ops * 100 if total_ops > 0 else 0
    
    if success_rate < 100:
        failed = total_ops - int(total_ops * success_rate / 100)
        analysis['issues'].append(f"Test failures detected: {failed} operations failed")
        analysis['performance_grade'] = 'B'
    
    # Check DynamoDB throttling (user errors)
    products_errors = metrics.get('products_table', {}).get('user_errors')
    carts_errors = metrics.get('carts_table', {}).get('user_errors')
    
    if products_errors and products_errors['max'] > 0:
        analysis['issues'].append(f"Products table throttling detected: {int(products_errors['max'])} errors")
        analysis['recommendations'].append("Consider increasing DynamoDB on-demand capacity or using provisioned capacity")
        analysis['performance_grade'] = 'C'
    
    if carts_errors and carts_errors['max'] > 0:
        analysis['issues'].append(f"Carts table throttling detected: {int(carts_errors['max'])} errors")
        analysis['recommendations'].append("Consider increasing DynamoDB on-demand capacity or using provisioned capacity")
        analysis['performance_grade'] = 'C'
    
    # Check ECS CPU
    ecs_cpu = metrics.get('ecs', {}).get('cpu')
    if ecs_cpu and ecs_cpu['avg'] > 70:
        analysis['issues'].append(f"High ECS CPU utilization: {ecs_cpu['avg']}%")
        analysis['recommendations'].append("Consider scaling ECS tasks or increasing CPU allocation")
        if analysis['performance_grade'] == 'A':
            analysis['performance_grade'] = 'B'
    
    # Check response times
    ops = stats.get('operations', {}) if isinstance(stats, dict) else {}
    for op_name, op_stats in ops.items():
        avg_time = op_stats.get('avg_response_time', 0) if isinstance(op_stats, dict) else 0
        if avg_time > 500:
            analysis['issues'].append(f"Slow {op_name} operations: {avg_time:.2f}ms avg")
            analysis['recommendations'].append(f"Investigate {op_name} performance - DynamoDB should be faster")
            if analysis['performance_grade'] == 'A':
                analysis['performance_grade'] = 'B'
    
    # Check DynamoDB latency
    products_latency = metrics.get('products_table', {}).get('getitem_latency')
    carts_put_latency = metrics.get('carts_table', {}).get('putitem_latency')
    carts_get_latency = metrics.get('carts_table', {}).get('getitem_latency')
    
    if products_latency and products_latency['avg'] > 20:
        analysis['issues'].append(f"High products table latency: {products_latency['avg']}ms")
        analysis['recommendations'].append("Review DynamoDB query patterns and consider using DAX for caching")
    
    if carts_put_latency and carts_put_latency['avg'] > 20:
        analysis['issues'].append(f"High carts write latency: {carts_put_latency['avg']}ms")
        analysis['recommendations'].append("Review data model and consider batching write operations")
    
    if carts_get_latency and carts_get_latency['avg'] > 20:
        analysis['issues'].append(f"High carts read latency: {carts_get_latency['avg']}ms")
        analysis['recommendations'].append("Consider using DynamoDB DAX for read caching")
    
    if not analysis['issues']:
        analysis['summary'] = "Excellent performance! All metrics within acceptable ranges for DynamoDB."
    else:
        analysis['summary'] = f"Found {len(analysis['issues'])} performance issues."
    
    return analysis

def print_report_summary(report):
    """Print human-readable report summary"""
    
    print("\n" + "="*70)
    print("COMPREHENSIVE DYNAMODB TEST REPORT")
    print("="*70)
    
    # Test Results
    test_stats = report['test_results']['statistics']
    print("\n📊 TEST RESULTS:")
    print("-"*70)
    
    # Calculate totals from results if statistics is missing keys
    results = report.get('test_results', {}).get('results', [])
    if results:
        total = len(results)
        successful = sum(1 for r in results if r.get('success', False))
        failed = total - successful
        success_rate = (successful / total * 100) if total > 0 else 0
    else:
        total = 0
        successful = 0
        failed = 0
        success_rate = 0
    
    print(f"Total Operations:     {total}")
    print(f"Successful:           {successful}")
    print(f"Failed:               {failed}")
    print(f"Success Rate:         {success_rate:.2f}%")
    
    # Response Times
    print("\n⏱️  RESPONSE TIMES:")
    print("-"*70)
    for op_name, op_stats in test_stats.items():
        if isinstance(op_stats, dict) and 'avg_response_time' in op_stats:
            print(f"{op_name:20s} avg: {op_stats.get('avg_response_time', 0):6.2f}ms  " +
                  f"(min: {op_stats.get('min_response_time', 0):6.2f}ms, " +
                  f"max: {op_stats.get('max_response_time', 0):6.2f}ms)")
    
    # DynamoDB Products Table Metrics
    products = report['cloudwatch_metrics']['products_table']
    print("\n📦 PRODUCTS TABLE METRICS:")
    print("-"*70)
    if products['read_capacity']:
        print(f"Read Capacity Used:  {products['read_capacity']['avg']:6.2f} units " +
              f"(total: {products['read_capacity']['max']:6.2f})")
    if products['write_capacity']:
        print(f"Write Capacity Used: {products['write_capacity']['avg']:6.2f} units " +
              f"(total: {products['write_capacity']['max']:6.2f})")
    if products['getitem_latency']:
        print(f"GetItem Latency:     {products['getitem_latency']['avg']:6.2f}ms " +
              f"(max: {products['getitem_latency']['max']:6.2f}ms)")
    if products['user_errors']:
        print(f"Throttle Errors:     {int(products['user_errors']['max'])}")
    
    # DynamoDB Carts Table Metrics
    carts = report['cloudwatch_metrics']['carts_table']
    print("\n🛒 CARTS TABLE METRICS:")
    print("-"*70)
    if carts['read_capacity']:
        print(f"Read Capacity Used:  {carts['read_capacity']['avg']:6.2f} units " +
              f"(total: {carts['read_capacity']['max']:6.2f})")
    if carts['write_capacity']:
        print(f"Write Capacity Used: {carts['write_capacity']['avg']:6.2f} units " +
              f"(total: {carts['write_capacity']['max']:6.2f})")
    if carts['putitem_latency']:
        print(f"PutItem Latency:     {carts['putitem_latency']['avg']:6.2f}ms " +
              f"(max: {carts['putitem_latency']['max']:6.2f}ms)")
    if carts['getitem_latency']:
        print(f"GetItem Latency:     {carts['getitem_latency']['avg']:6.2f}ms " +
              f"(max: {carts['getitem_latency']['max']:6.2f}ms)")
    if carts['user_errors']:
        print(f"Throttle Errors:     {int(carts['user_errors']['max'])}")
    
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
    print("Report saved to: comprehensive_dynamodb_report.json")
    print("="*70)
    print("\nTo compare with MySQL:")
    print("  cat comprehensive_dynamodb_report.json | jq '.performance_grade'")
    print("  cat comprehensive_report.json | jq '.performance_grade'")
    print("="*70)

if __name__ == "__main__":
    generate_report()