variable "service_name" {
  description = "Name of the service (used for tagging)"
  type        = string
}

variable "products_table_name" {
  description = "Name of the DynamoDB products table"
  type        = string
  default     = "ecommerce-products"
}

variable "carts_table_name" {
  description = "Name of the DynamoDB carts table"
  type        = string
  default     = "ecommerce-carts"
}