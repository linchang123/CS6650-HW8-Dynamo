output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.alb.alb_dns_name
}

output "application_url" {
  description = "URL to access your application"
  value       = "http://${module.alb.alb_dns_name}"
}

output "dynamodb_products_table" {
  description = "Name of the DynamoDB products table"
  value       = module.dynamodb.products_table_name
}

output "dynamodb_carts_table" {
  description = "Name of the DynamoDB carts table"
  value       = module.dynamodb.carts_table_name
}