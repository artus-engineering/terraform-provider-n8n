variable "n8n_host" {
  description = "The n8n instance host URL"
  type        = string
  default     = "http://localhost:5678"
}

variable "n8n_api_key" {
  description = "The API key for authenticating with n8n"
  type        = string
  sensitive   = true
}

