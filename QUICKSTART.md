# Quick Start Guide

This guide will help you get started with the n8n Terraform provider quickly.

## Prerequisites

1. **Terraform** (>= 1.0) installed
2. **n8n instance** with API access enabled
3. **n8n API Key** - You can generate this in your n8n instance settings

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/artus-engineering/terraform-provider-n8n.git
cd terraform-provider-n8n

# Build the provider
make build

# Install locally
make install
```

### Option 2: Use Development Override

Add to your `~/.terraformrc`:

```
provider_installation {
  dev_overrides {
    "artus-engineering/n8n" = "/path/to/terraform-provider-n8n"
  }
  direct {}
}
```

## Basic Usage

1. **Create a Terraform configuration file** (`main.tf`):

```hcl
terraform {
  required_providers {
    n8n = {
      source  = "artus-engineering/n8n"
      version = "~> 1.0"
    }
  }
}

provider "n8n" {
  host    = "https://your-n8n-instance.com"
  api_key = "your-api-key-here"
}

resource "n8n_credential" "example" {
  name = "my-http-basic-auth"
  type = "httpBasicAuth"
  data = jsonencode({
    user     = "username"
    password = "password"
  })
}
```

2. **Initialize Terraform**:

```bash
terraform init
```

3. **Plan the changes**:

```bash
terraform plan
```

4. **Apply the configuration**:

```bash
terraform apply
```

## Example: Creating Multiple Credentials

```hcl
# HTTP Basic Auth
resource "n8n_credential" "http_basic" {
  name = "api-basic-auth"
  type = "httpBasicAuth"
  data = jsonencode({
    user     = "api_user"
    password = "secure_password"
  })
}

# OAuth2
resource "n8n_credential" "oauth2" {
  name = "oauth2-credential"
  type = "oAuth2Api"
  data = jsonencode({
    clientId                  = "client-id"
    clientSecret              = "client-secret"
    accessTokenUrl            = "https://api.example.com/oauth/token"
    authUrl                   = "https://api.example.com/oauth/authorize"
    scope                     = "read write"
    authQueryParameters       = ""
    sendAdditionalBodyProperties = false
    additionalBodyProperties  = ""
  })
  nodes_access = ["httpRequest"]
}

# HTTP Header Auth
resource "n8n_credential" "header_auth" {
  name = "bearer-token"
  type = "httpHeaderAuth"
  data = jsonencode({
    name  = "Authorization"
    value = "Bearer your-token-here"
  })
}
```

## Environment Variables

Instead of hardcoding credentials, use environment variables:

```hcl
provider "n8n" {
  host    = var.n8n_host
  api_key = var.n8n_api_key
}
```

Set them before running Terraform:

```bash
export N8N_HOST="https://your-n8n-instance.com"
export N8N_API_KEY="your-api-key"
terraform apply
```

Or use a `.tfvars` file:

```hcl
# terraform.tfvars
n8n_host   = "https://your-n8n-instance.com"
n8n_api_key = "your-api-key"
```

## Troubleshooting

### Error: "Missing n8n API Host"

Make sure you've set the `host` parameter in your provider configuration or the `N8N_HOST` environment variable.

### Error: "Missing n8n API Key"

Ensure you've provided the `api_key` parameter or set the `N8N_API_KEY` environment variable.

### Error: "API error (status 401)"

Your API key is invalid or expired. Generate a new one in your n8n instance settings.

### Error: "API error (status 404)"

The n8n API endpoint might be incorrect, or your n8n instance doesn't have the API enabled. Check your n8n configuration.

## Next Steps

- Check out the [examples](./examples/) directory for more use cases
- Read the full [README](./README.md) for detailed documentation
- Review the [n8n API documentation](https://docs.n8n.io/api/) for available credential types
