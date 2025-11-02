output "products_table_name" {
  description = "Name of the products DynamoDB table"
  value       = aws_dynamodb_table.products.name
}

output "products_table_arn" {
  description = "ARN of the products DynamoDB table"
  value       = aws_dynamodb_table.products.arn
}

output "carts_table_name" {
  description = "Name of the carts DynamoDB table"
  value       = aws_dynamodb_table.carts.name
}

output "carts_table_arn" {
  description = "ARN of the carts DynamoDB table"
  value       = aws_dynamodb_table.carts.arn
}