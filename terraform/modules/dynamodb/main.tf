# DynamoDB table for products
resource "aws_dynamodb_table" "products" {
  name           = var.products_table_name
  billing_mode   = "PAY_PER_REQUEST"  # On-demand billing (no capacity planning needed)
  hash_key       = "product_id"

  attribute {
    name = "product_id"
    type = "N"  # Number type
  }

  tags = {
    Name        = var.products_table_name
    Environment = "dev"
    Service     = var.service_name
  }
}

# DynamoDB table for shopping carts
resource "aws_dynamodb_table" "carts" {
  name           = var.carts_table_name
  billing_mode   = "PAY_PER_REQUEST"  # On-demand billing
  hash_key       = "customer_id"

  attribute {
    name = "customer_id"
    type = "N"  # Number type
  }

  tags = {
    Name        = var.carts_table_name
    Environment = "dev"
    Service     = var.service_name
  }
}