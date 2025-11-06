terraform {
  required_providers {
    n8n = {
      source = "artus-engineering/n8n"
    }
  }
}

provider "n8n" {
  host    = var.n8n_host
  api_key = var.n8n_api_key
}

# Example: HTTP Basic Auth credential
resource "n8n_credential" "http_basic" {
  name = "example-http-basic-auth"

  basic_auth {
    username = "myusername"
    password = "mypassword"
  }
}

# Example: OAuth2 API credential
resource "n8n_credential" "oauth2" {
  name = "example-oauth2"

  oauth2 {
    client_id        = "your-client-id"
    client_secret    = "your-client-secret"
    access_token_url = "https://example.com/oauth/token"
    auth_url         = "https://example.com/oauth/authorize"
    scope            = "read write"
  }

  nodes_access = ["httpRequest"]
}

# Example: HTTP Header Auth credential
resource "n8n_credential" "http_header" {
  name = "example-http-header"

  header_auth {
    name  = "Authorization"
    value = "Bearer your-token-here"
  }
}
