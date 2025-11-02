variable "service_name" {
  type        = string
  description = "Name of the service"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID"
}

variable "subnet_ids" {
  type        = list(string)
  description = "List of subnet IDs for ALB (must be in at least 2 AZs)"
}

variable "container_port" {
  type        = number
  description = "Port the container listens on"
}