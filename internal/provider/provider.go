package provider

import (
	"context"

	"github.com/artus-engineering/terraform-provider-n8n/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &n8nProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &n8nProvider{
			version: version,
		}
	}
}

// n8nProvider is the provider implementation.
type n8nProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// n8nProviderModel maps provider schema data to a Go type.
type n8nProviderModel struct {
	Host     types.String `tfsdk:"host"`
	APIKey   types.String `tfsdk:"api_key"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// Metadata returns the provider type name.
func (p *n8nProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "n8n"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *n8nProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with n8n API to manage credentials and other resources.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The n8n instance host URL (e.g., https://n8n.example.com).",
				Required:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for authenticating with n8n.",
				Required:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Allow insecure HTTPS connections. Defaults to false.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a n8n API client for data sources and resources.
func (p *n8nProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring n8n client")

	var config n8nProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown n8n API Host",
			"The provider cannot create the n8n API client as there is an unknown configuration value for the n8n API host. "+
				"Either apply the source of the value first, or set the value statically in the configuration.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown n8n API Key",
			"The provider cannot create the n8n API client as there is an unknown configuration value for the n8n API key. "+
				"Either apply the source of the value first, or set the value statically in the configuration.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Get required values (they are required in schema, so they should be present)
	host := config.Host.ValueString()
	apiKey := config.APIKey.ValueString()

	// Get optional insecure value
	insecure := false
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	// Validate that required values are not empty
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing n8n API Host",
			"The provider cannot create the n8n API client as there is an empty value for the n8n API host. "+
				"Ensure the host value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing n8n API Key",
			"The provider cannot create the n8n API client as there is an empty value for the n8n API key. "+
				"Ensure the api_key value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "n8n_host", host)
	ctx = tflog.SetField(ctx, "n8n_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "n8n_api_key")

	tflog.Debug(ctx, "Creating n8n client")

	// Create a new n8n client using the configuration values
	n8nClient, err := client.NewClient(&host, &apiKey, &insecure)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create n8n API Client",
			"An unexpected error occurred when creating the n8n API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"n8n Client Error: "+err.Error(),
		)
		return
	}

	// Make the n8n client available during DataSource and Resource
	// type Configure methods.
	resp.ResourceData = n8nClient
	resp.DataSourceData = n8nClient

	tflog.Info(ctx, "Configured n8n client", map[string]any{"success": true})
}

// Resources defines the provider resources.
func (p *n8nProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCredentialResource,
	}
}

// DataSources defines the provider data sources.
func (p *n8nProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewCredentialDataSource,
	}
}
