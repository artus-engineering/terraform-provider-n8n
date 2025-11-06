package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	apiVersion     = "v1"
)

// Client handles communication with the n8n API.
type Client struct {
	Host     string
	APIKey   string
	Insecure bool
	client   *http.Client
}

// NewClient creates a new n8n API client.
func NewClient(host, apiKey *string, insecure *bool) (*Client, error) {
	if host == nil || *host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if apiKey == nil || *apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure != nil && *insecure,
		},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   defaultTimeout,
	}

	return &Client{
		Host:     *host,
		APIKey:   *apiKey,
		Insecure: insecure != nil && *insecure,
		client:   httpClient,
	}, nil
}

// doRequest performs an HTTP request to the n8n API.
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/api/%s/%s", c.Host, apiVersion, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Credential represents an n8n credential.
type Credential struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	NodesAccess []NodeAccess           `json:"nodesAccess,omitempty"`
}

// NodeAccess defines which nodes can access the credential.
type NodeAccess struct {
	NodeType string `json:"nodeType"`
}

// CreateCredential creates a new credential in n8n.
func (c *Client) CreateCredential(credential *Credential) (*Credential, error) {
	body := map[string]interface{}{
		"name": credential.Name,
		"type": credential.Type,
		"data": credential.Data,
	}

	if len(credential.NodesAccess) > 0 {
		body["nodesAccess"] = credential.NodesAccess
	}

	respBody, err := c.doRequest("POST", "credentials", body)
	if err != nil {
		return nil, err
	}

	var createdCredential Credential
	if err := json.Unmarshal(respBody, &createdCredential); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &createdCredential, nil
}

// ListCredentialsResponse represents the response from listing credentials.
type ListCredentialsResponse struct {
	Data []Credential `json:"data"`
}

// ListCredentials retrieves all credentials.
func (c *Client) ListCredentials() ([]Credential, error) {
	respBody, err := c.doRequest("GET", "credentials", nil)
	if err != nil {
		return nil, err
	}

	var response ListCredentialsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		// Try to unmarshal as a direct array if the response doesn't have a "data" wrapper
		var credentials []Credential
		if err2 := json.Unmarshal(respBody, &credentials); err2 != nil {
			return nil, fmt.Errorf("error unmarshaling response: %w", err)
		}
		return credentials, nil
	}

	return response.Data, nil
}

// GetCredential retrieves a credential by ID.
// Since n8n API may not support direct GET by ID, we list all credentials and find the matching one.
func (c *Client) GetCredential(id string) (*Credential, error) {
	// First, try direct GET (in case the API supports it)
	respBody, err := c.doRequest("GET", fmt.Sprintf("credentials/%s", id), nil)
	if err == nil {
		var credential Credential
		if err := json.Unmarshal(respBody, &credential); err != nil {
			return nil, fmt.Errorf("error unmarshaling response: %w", err)
		}
		return &credential, nil
	}

	// If direct GET fails, fall back to listing and filtering
	credentials, err := c.ListCredentials()
	if err != nil {
		return nil, fmt.Errorf("error listing credentials: %w", err)
	}

	for _, cred := range credentials {
		if cred.ID == id {
			return &cred, nil
		}
	}

	return nil, fmt.Errorf("credential with ID %s not found", id)
}

// UpdateCredential updates an existing credential by deleting and recreating it.
// Note: The n8n API does not support PUT or PATCH for credentials, so we must
// delete and recreate. This will result in a new credential ID.
// WARNING: If workflows reference this credential by ID, they will need to be updated.
func (c *Client) UpdateCredential(id string, credential *Credential) (*Credential, error) {
	// Delete the old credential
	err := c.DeleteCredential(id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old credential before update: %w", err)
	}

	// Create a new credential with the updated data
	// This will generate a new ID
	newCredential, err := c.CreateCredential(credential)
	if err != nil {
		return nil, fmt.Errorf("failed to create new credential after delete: %w", err)
	}

	return newCredential, nil
}

// DeleteCredential deletes a credential by ID.
func (c *Client) DeleteCredential(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("credentials/%s", id), nil)
	return err
}
