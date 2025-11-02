variable "service_name" {
  type        = string
  description = "Base name for ECS resources"
}

variable "image" {
  type        = string
  description = "ECR image URI (with tag)"
}

variable "container_port" {
  type        = number
  description = "Port your app listens on"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnets for FARGATE tasks"
}

variable "security_group_ids" {
  type        = list(string)
  description = "SGs for FARGATE tasks"
}

variable "execution_role_arn" {
  type        = string
  description = "ECS Task Execution Role ARN"
}

variable "task_role_arn" {
  type        = string
  description = "IAM Role ARN for app permissions"
}

variable "log_group_name" {
  type        = string
  description = "CloudWatch log group name"
}

variable "ecs_count" {
  type        = number
  default     = 1
  description = "Desired Fargate task count"
}

variable "region" {
  type        = string
  description = "AWS region (for awslogs driver)"
}

variable "cpu" {
  type        = string
  default     = "256"
  description = "vCPU units"
}

variable "memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB)"
}


# ALB variables
variable "target_group_arn" {
  type        = string
  description = "ALB Target Group ARN"
}

# Auto-scaling variables
variable "min_capacity" {
  type        = number
  default     = 1
  description = "Minimum number of tasks"
}

variable "max_capacity" {
  type        = number
  default     = 10
  description = "Maximum number of tasks"
}


# DynamoDB table names
variable "products_table_name" {
  description = "Name of the DynamoDB products table"
  type        = string
}

variable "carts_table_name" {
  description = "Name of the DynamoDB carts table"
  type        = string
}