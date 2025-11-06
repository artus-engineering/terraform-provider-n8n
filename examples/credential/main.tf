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
  name = "example-http-basic-auth-awfw"
  type = "httpBasicAuth"
  data = jsonencode({
    user     = "myusername1"
    password = "mypassword"
  })
}

# Example: OAuth2 API credential
resource "n8n_credential" "oauth2" {
  name = "example-oauth2"
  type = "oAuth2Api"
  data = jsonencode({
    clientId                     = "your-client-id"
    clientSecret                 = "your-client-secret"
    accessTokenUrl               = "https://example.com/oauth/token"
    authUrl                      = "https://example.com/oauth/authorize"
    scope                        = "read write"
    authQueryParameters          = ""
    sendAdditionalBodyProperties = false
    additionalBodyProperties     = ""
  })

  nodes_access = ["httpRequest"]
}

# Example: HTTP Header Auth credential
resource "n8n_credential" "http_header" {
  name = "example-http-header"
  type = "httpHeaderAuth"
  data = jsonencode({
    name  = "Authorization"
    value = "Bearer your-token-here"
  })
}

