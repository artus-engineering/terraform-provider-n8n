package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		host      *string
		apiKey    *string
		insecure  *bool
		wantError bool
	}{
		{
			name:      "valid client",
			host:      stringPtr("https://n8n.example.com"),
			apiKey:    stringPtr("test-api-key"),
			insecure:  boolPtr(false),
			wantError: false,
		},
		{
			name:      "missing host",
			host:      nil,
			apiKey:    stringPtr("test-api-key"),
			insecure:  boolPtr(false),
			wantError: true,
		},
		{
			name:      "empty host",
			host:      stringPtr(""),
			apiKey:    stringPtr("test-api-key"),
			insecure:  boolPtr(false),
			wantError: true,
		},
		{
			name:      "missing api key",
			host:      stringPtr("https://n8n.example.com"),
			apiKey:    nil,
			insecure:  boolPtr(false),
			wantError: true,
		},
		{
			name:      "empty api key",
			host:      stringPtr("https://n8n.example.com"),
			apiKey:    stringPtr(""),
			insecure:  boolPtr(false),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.host, tt.apiKey, tt.insecure)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if client != nil {
					t.Errorf("Expected nil client but got %+v", client)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Errorf("Expected client but got nil")
				}
				if client != nil && client.Host != *tt.host {
					t.Errorf("Expected host %s, got %s", *tt.host, client.Host)
				}
				if client != nil && client.APIKey != *tt.apiKey {
					t.Errorf("Expected API key %s, got %s", *tt.apiKey, client.APIKey)
				}
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
