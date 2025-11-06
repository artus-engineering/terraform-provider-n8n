# Terraform Provider for n8n

[![Test](https://github.com/artus-engineering/terraform-provider-n8n/actions/workflows/test.yml/badge.svg)](https://github.com/artus-engineering/terraform-provider-n8n/actions/workflows/test.yml)

A Terraform provider for managing n8n resources, starting with credential management.

## Features

> **Note**: This is an early and minimal provider. It does not support all n8n resources but is under development and will be extended step by step.

- **Credential Management**: Manage n8n credentials

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- n8n instance with API access enabled
- n8n API Key (can be generated in your n8n instance settings)

## Quick Start

### 1. Install the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    n8n = {
      source  = "artus-engineering/n8n"
    }
  }
}
```

### 2. Configure the Provider

```hcl
provider "n8n" {
  host    = "https://your-n8n-instance.com"
  api_key = "your-api-key-here"
}
```

### 3. Create Your First Credential

```hcl
resource "n8n_credential" "example" {
  name = "my-http-basic-auth"
  type = "httpBasicAuth"
  data = jsonencode({
    user     = "username"
    password = "password"
  })
}
```

### 4. Initialize and Apply

```bash
terraform init
terraform plan
terraform apply
```

## Installation

### From Terraform Registry (Recommended)

Once published, the provider will be available directly from the Terraform Registry. Simply add it to your `required_providers` block as shown in the Quick Start section.

### Building from Source

If you need to build from source:

1. Clone the repository:
```bash
git clone https://github.com/artus-engineering/terraform-provider-n8n.git
cd terraform-provider-n8n
```

2. Build the provider:
```bash
make build
```

3. Install locally:
```bash
make install
```

### Development Override

For local development, you can use a development override in your `~/.terraformrc`:

```
provider_installation {
  dev_overrides {
    "artus-engineering/n8n" = "/path/to/terraform-provider-n8n"
  }
  direct {}
}
```

## Configuration

The provider requires the following configuration:

```hcl
provider "n8n" {
  host    = "https://n8n.example.com"  # Your n8n instance URL
  api_key = "your-api-key-here"        # Your n8n API key
  insecure = false                      # Optional: Allow insecure HTTPS (default: false)
}
```

### Using Variables

Instead of hardcoding credentials, use Terraform variables:

```hcl
variable "n8n_host" {
  description = "n8n instance URL"
  type        = string
}

variable "n8n_api_key" {
  description = "n8n API key"
  type        = string
  sensitive   = true
}

provider "n8n" {
  host    = var.n8n_host
  api_key = var.n8n_api_key
}
```

Set them via a `.tfvars` file:

```hcl
# terraform.tfvars
n8n_host   = "https://your-n8n-instance.com"
n8n_api_key = "your-api-key"
```

## Usage

### Managing Credentials

#### HTTP Basic Auth Credential

```hcl
resource "n8n_credential" "http_basic" {
  name = "my-http-basic-auth"
  type = "httpBasicAuth"
  data = jsonencode({
    user     = "myusername"
    password = "mypassword"
  })
}
```

#### OAuth2 API Credential

```hcl
resource "n8n_credential" "oauth2" {
  name = "my-oauth2-credential"
  type = "oAuth2Api"
  data = jsonencode({
    clientId                  = "your-client-id"
    clientSecret              = "your-client-secret"
    accessTokenUrl            = "https://example.com/oauth/token"
    authUrl                   = "https://example.com/oauth/authorize"
    scope                     = "read write"
    authQueryParameters       = ""
    sendAdditionalBodyProperties = false
    additionalBodyProperties  = ""
  })

  nodes_access = ["httpRequest"]
}
```

#### HTTP Header Auth Credential

```hcl
resource "n8n_credential" "http_header" {
  name = "my-http-header-auth"
  type = "httpHeaderAuth"
  data = jsonencode({
    name  = "Authorization"
    value = "Bearer your-token-here"
  })
}
```

### Creating Multiple Credentials

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

## Resources

### `n8n_credential`

Manages a credential in n8n.

**Important Note on Updates**: The n8n API does not support updating credentials directly (no PUT/PATCH endpoints). When you update a credential, the provider will delete the old credential and create a new one. This means:
- The credential `id` will change after an update
- If workflows reference this credential by ID, they will need to be updated to use the new ID
- There will be a brief moment during the update where the credential doesn't exist

#### Arguments

- `name` (Required, string): The name of the credential
- `type` (Required, string): The type of credential (e.g., `httpBasicAuth`, `httpHeaderAuth`, `oAuth2Api`, etc.)
- `data` (Required, string, sensitive): JSON string containing the credential data
- `nodes_access` (Optional, list of strings): List of node types that can access this credential

#### Attributes

- `id` (Computed, string): The unique identifier of the credential. **Note**: This ID will change when the credential is updated.

#### Example

```hcl
resource "n8n_credential" "example" {
  name = "example-credential"
  type = "httpBasicAuth"
  data = jsonencode({
    user     = "admin"
    password = "secret"
  })
}
```

## Troubleshooting

### Error: "Missing n8n API Host"

Make sure you've set the `host` parameter in your provider configuration.

### Error: "Missing n8n API Key"

Ensure you've provided the `api_key` parameter in your provider configuration.

### Error: "API error (status 401)"

Your API key is invalid or expired. Generate a new one in your n8n instance settings.

### Error: "API error (status 404)"

The n8n API endpoint might be incorrect, or your n8n instance doesn't have the API enabled. Check your n8n configuration.

## Development

### Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- Make (for using the Makefile)

### Building

```bash
make build
```

### Testing

Run unit tests:
```bash
make test
```

Run acceptance tests (requires a running n8n instance):
```bash
make testacc
```

### Linting

```bash
make lint
```

### Formatting

```bash
make fmt
```

### Running All Checks

```bash
make check
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## References

- [n8n API Documentation](https://docs.n8n.io/api/)
- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- Check out the [examples](./examples/) directory for more use cases
