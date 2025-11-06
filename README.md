# Terraform Provider for n8n

[![Test](https://github.com/artus-engineering/terraform-provider-n8n/actions/workflows/test.yml/badge.svg)](https://github.com/artus-engineering/terraform-provider-n8n/actions/workflows/test.yml)

A Terraform provider for managing n8n resources, starting with credential management.

## Features

- **Credential Management**: Create, read, update, and delete n8n credentials
- **Secure**: Credential data is marked as sensitive and handled securely
- **Modern**: Built with Terraform Plugin Framework (latest best practices)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)
- n8n instance with API access enabled

## Installation

### Using Terraform CLI Configuration File

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    n8n = {
      source  = "artus-engineering/n8n"
      version = "~> 1.0"
    }
  }
}
```

### Building from Source

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

## Configuration

The provider requires the following configuration:

```hcl
provider "n8n" {
  host    = "https://n8n.example.com"  # Your n8n instance URL
  api_key = "your-api-key-here"        # Your n8n API key
  insecure = false                      # Optional: Allow insecure HTTPS (default: false)
}
```

### Environment Variables

You can also configure the provider using environment variables:

- `N8N_HOST`: The n8n instance host URL
- `N8N_API_KEY`: The API key for authenticating with n8n

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
