package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestCredentialResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	NewCredentialResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	validateSchemaAttributeExists(t, schemaResponse.Schema, "id")
	validateSchemaAttributeExists(t, schemaResponse.Schema, "name")
	validateSchemaAttributeExists(t, schemaResponse.Schema, "nodes_access")

	// Validate blocks exist
	if _, ok := schemaResponse.Schema.Blocks["basic_auth"]; !ok {
		t.Errorf("missing block: basic_auth")
	}
	if _, ok := schemaResponse.Schema.Blocks["oauth2"]; !ok {
		t.Errorf("missing block: oauth2")
	}
	if _, ok := schemaResponse.Schema.Blocks["header_auth"]; !ok {
		t.Errorf("missing block: header_auth")
	}
}

func validateSchemaAttributeExists(t *testing.T, s schema.Schema, attributeName string) {
	t.Helper()

	attribute, ok := s.Attributes[attributeName]
	if !ok {
		t.Errorf("missing attribute: %s", attributeName)
		return
	}

	if attribute == nil {
		t.Errorf("attribute is nil: %s", attributeName)
	}
}

func TestCredentialResourceMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadataRequest := resource.MetadataRequest{
		ProviderTypeName: "n8n",
	}
	metadataResponse := &resource.MetadataResponse{}

	NewCredentialResource().Metadata(ctx, metadataRequest, metadataResponse)

	if metadataResponse.TypeName != "n8n_credential" {
		t.Errorf("Expected TypeName to be 'n8n_credential', got '%s'", metadataResponse.TypeName)
	}
}
