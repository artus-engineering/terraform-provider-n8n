package provider

import (
	"context"
	"fmt"

	"github.com/artus-engineering/terraform-provider-n8n/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &credentialResource{}
	_ resource.ResourceWithConfigure   = &credentialResource{}
	_ resource.ResourceWithImportState = &credentialResource{}
	_ resource.ResourceWithModifyPlan  = &credentialResource{}
)

// NewCredentialResource is a helper function to simplify the provider implementation.
func NewCredentialResource() resource.Resource {
	return &credentialResource{}
}

// credentialResource is the resource implementation.
type credentialResource struct {
	client *client.Client
}

// credentialResourceModel maps the resource schema data.
type credentialResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	BasicAuth   types.Object `tfsdk:"basic_auth"`
	OAuth2      types.Object `tfsdk:"oauth2"`
	HeaderAuth  types.Object `tfsdk:"header_auth"`
	NodesAccess types.List   `tfsdk:"nodes_access"`
}

// basicAuthModel represents the httpBasicAuth credential block.
type basicAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// oAuth2Model represents the oAuth2Api credential block.
type oAuth2Model struct {
	ClientId                     types.String `tfsdk:"client_id"`
	ClientSecret                 types.String `tfsdk:"client_secret"`
	AccessTokenUrl               types.String `tfsdk:"access_token_url"`
	AuthUrl                      types.String `tfsdk:"auth_url"`
	Scope                        types.String `tfsdk:"scope"`
	AuthQueryParameters          types.String `tfsdk:"auth_query_parameters"`
	SendAdditionalBodyProperties types.Bool   `tfsdk:"send_additional_body_properties"`
	AdditionalBodyProperties     types.String `tfsdk:"additional_body_properties"`
}

// headerAuthModel represents the httpHeaderAuth credential block.
type headerAuthModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// Metadata returns the resource type name.
func (r *credentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

// Schema defines the schema for the resource.
func (r *credentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a credential in n8n. Credentials are used to authenticate with external services. Exactly one credential type block must be specified.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the credential.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the credential.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nodes_access": schema.ListAttribute{
				Description: "List of node types that can access this credential. Each item should be a string representing the node type.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					&requiresReplaceListModifier{},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"basic_auth": schema.SingleNestedBlock{
				Description: "HTTP Basic Authentication credentials.",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "The username for basic authentication.",
						Optional:    true, // Made optional - validated in ModifyPlan
						Sensitive:   false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"password": schema.StringAttribute{
						Description: "The password for basic authentication.",
						Optional:    true, // Made optional - validated in ModifyPlan
						Sensitive:   true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					&requiresReplaceObjectModifier{},
				},
			},
			"oauth2": schema.SingleNestedBlock{
				Description: "OAuth2 API credentials.",
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Description: "The OAuth2 client ID.",
						Optional:    true, // Made optional - validated in ModifyPlan
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"client_secret": schema.StringAttribute{
						Description: "The OAuth2 client secret.",
						Optional:    true, // Made optional - validated in ModifyPlan
						Sensitive:   true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"access_token_url": schema.StringAttribute{
						Description: "The URL to obtain the access token.",
						Optional:    true, // Made optional - validated in ModifyPlan
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"auth_url": schema.StringAttribute{
						Description: "The OAuth2 authorization URL.",
						Optional:    true, // Made optional - validated in ModifyPlan
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"scope": schema.StringAttribute{
						Description: "The OAuth2 scope.",
						Optional:    true, // Made optional - validated in ModifyPlan
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"auth_query_parameters": schema.StringAttribute{
						Description: "Additional query parameters for the authorization request.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"send_additional_body_properties": schema.BoolAttribute{
						Description: "Whether to send additional body properties.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							&requiresReplaceBoolModifier{},
						},
					},
					"additional_body_properties": schema.StringAttribute{
						Description: "Additional body properties to send.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					&requiresReplaceObjectModifier{},
				},
			},
			"header_auth": schema.SingleNestedBlock{
				Description: "HTTP Header Authentication credentials.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "The header name (e.g., 'Authorization').",
						Optional:    true, // Made optional - validated in ModifyPlan
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"value": schema.StringAttribute{
						Description: "The header value (e.g., 'Bearer token').",
						Optional:    true, // Made optional - validated in ModifyPlan
						Sensitive:   true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					&requiresReplaceObjectModifier{},
				},
			},
		},
	}

	// Set ExactlyOneOf validation using custom validation
	// Note: Terraform Plugin Framework doesn't have built-in ExactlyOneOf for blocks,
	// so we'll validate this in the Create/Update methods
}

// Configure adds the provider configured client to the resource.
func (r *credentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *credentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan credentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one credential block is defined and extract type/data
	credentialType, data, err := validateCredentialBlocks(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Credential Configuration",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Creating credential", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"type": credentialType,
	})

	// Convert nodes_access to []client.NodeAccess
	var nodesAccess []client.NodeAccess
	if !plan.NodesAccess.IsNull() && !plan.NodesAccess.IsUnknown() {
		var nodeTypes []types.String
		diags := plan.NodesAccess.ElementsAs(ctx, &nodeTypes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, nodeType := range nodeTypes {
			nodesAccess = append(nodesAccess, client.NodeAccess{
				NodeType: nodeType.ValueString(),
			})
		}
	}

	// Create the credential
	credential := &client.Credential{
		Name:        plan.Name.ValueString(),
		Type:        credentialType,
		Data:        data,
		NodesAccess: nodesAccess,
	}

	createdCredential, err := r.client.CreateCredential(credential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating credential",
			fmt.Sprintf("Could not create credential, unexpected error: %s", err.Error()),
		)
		return
	}

	// Map response body to resource schema attributes
	plan.ID = types.StringValue(createdCredential.ID)
	plan.Name = types.StringValue(createdCredential.Name)

	// Set nodes_access if it was provided
	if len(createdCredential.NodesAccess) > 0 {
		nodeTypeValues := make([]types.String, len(createdCredential.NodesAccess))
		for i, na := range createdCredential.NodesAccess {
			nodeTypeValues[i] = types.StringValue(na.NodeType)
		}
		nodesAccessList, diags := types.ListValueFrom(ctx, types.StringType, nodeTypeValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.NodesAccess = nodesAccessList
	}
	// Note: If nodesAccess was not provided in the response and was null in plan,
	// it will remain null, which is correct behavior

	// Note: We don't update the credential blocks from the response because n8n API
	// doesn't return sensitive credential data for security reasons.
	// The blocks remain as provided by the user.

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Created credential", map[string]interface{}{
		"id":   createdCredential.ID,
		"name": createdCredential.Name,
	})
}

// Read refreshes the Terraform state with the latest data.
// Note: n8n API may not support reading credentials for security reasons.
// If reading fails, we keep the existing state to avoid breaking Terraform operations.
func (r *credentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state credentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading credential", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	credential, err := r.client.GetCredential(state.ID.ValueString())
	if err != nil {
		// n8n API may not support reading credentials (security feature).
		// Instead of failing, we log a warning and keep the existing state.
		// This allows Terraform to continue working even if the API doesn't
		// support credential retrieval.
		tflog.Warn(ctx, "Could not read credential from API, keeping existing state", map[string]interface{}{
			"id":    state.ID.ValueString(),
			"error": err.Error(),
		})

		// Keep the existing state - don't update anything
		// The credential data is sensitive and n8n doesn't return it anyway,
		// so we preserve what we have in state.
		diags = resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update state with refreshed values (if we successfully read the credential)
	state.ID = types.StringValue(credential.ID)
	state.Name = types.StringValue(credential.Name)
	// Note: We don't update the credential blocks from the API response because
	// n8n doesn't return sensitive credential data. We keep the existing blocks.

	// Update nodes_access if present
	if len(credential.NodesAccess) > 0 {
		nodeTypeValues := make([]types.String, len(credential.NodesAccess))
		for i, na := range credential.NodesAccess {
			nodeTypeValues[i] = types.StringValue(na.NodeType)
		}
		nodesAccessList, diags := types.ListValueFrom(ctx, types.StringType, nodeTypeValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.NodesAccess = nodesAccessList
	} else {
		state.NodesAccess = types.ListNull(types.StringType)
	}

	// Note: The data field is not updated from the API response because
	// n8n doesn't return sensitive credential data. We keep the existing
	// value in state.

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read credential", map[string]interface{}{
		"id":   credential.ID,
		"name": credential.Name,
	})
}

// Update updates the resource and sets the updated Terraform state on success.
// Note: Updates are handled via replacement (delete and recreate) due to n8n API limitations.
func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan credentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one credential block is defined and extract type/data
	credentialType, data, err := validateCredentialBlocks(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Credential Configuration",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Updating credential via delete-and-recreate", map[string]interface{}{
		"old_id": plan.ID.ValueString(),
		"name":   plan.Name.ValueString(),
		"type":   credentialType,
	})

	// Convert nodes_access to []client.NodeAccess
	var nodesAccess []client.NodeAccess
	if !plan.NodesAccess.IsNull() && !plan.NodesAccess.IsUnknown() {
		var nodeTypes []types.String
		diags := plan.NodesAccess.ElementsAs(ctx, &nodeTypes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, nodeType := range nodeTypes {
			nodesAccess = append(nodesAccess, client.NodeAccess{
				NodeType: nodeType.ValueString(),
			})
		}
	}

	// Update the credential
	credential := &client.Credential{
		Name:        plan.Name.ValueString(),
		Type:        credentialType,
		Data:        data,
		NodesAccess: nodesAccess,
	}

	// Update credential by deleting and recreating (n8n API doesn't support PUT/PATCH)
	// Note: This will result in a new credential ID
	updatedCredential, err := r.client.UpdateCredential(plan.ID.ValueString(), credential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating credential",
			fmt.Sprintf("Could not update credential ID %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	// Log that the ID has changed
	if updatedCredential.ID != plan.ID.ValueString() {
		tflog.Info(ctx, "Credential ID changed after update", map[string]interface{}{
			"old_id": plan.ID.ValueString(),
			"new_id": updatedCredential.ID,
		})
	}

	// Map response body to resource schema attributes
	plan.ID = types.StringValue(updatedCredential.ID)
	plan.Name = types.StringValue(updatedCredential.Name)

	// Update nodes_access if it was provided
	if len(updatedCredential.NodesAccess) > 0 {
		nodeTypeValues := make([]types.String, len(updatedCredential.NodesAccess))
		for i, na := range updatedCredential.NodesAccess {
			nodeTypeValues[i] = types.StringValue(na.NodeType)
		}
		nodesAccessList, diags := types.ListValueFrom(ctx, types.StringType, nodeTypeValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.NodesAccess = nodesAccessList
	}
	// Note: If nodesAccess was not provided in the response and was null in plan,
	// it will remain null, which is correct behavior

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updated credential", map[string]interface{}{
		"id":   updatedCredential.ID,
		"name": updatedCredential.Name,
	})
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *credentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state credentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting credential", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	err := r.client.DeleteCredential(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting credential",
			fmt.Sprintf("Could not delete credential ID %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Deleted credential", map[string]interface{}{
		"id": state.ID.ValueString(),
	})
}

// ImportState imports the resource.
func (r *credentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ModifyPlan validates that exactly one credential block is provided.
// This runs during plan time to validate the configuration before Terraform
// validates nested block attributes.
func (r *credentialResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip validation during destroy or if plan is null
	if req.Plan.Raw.IsNull() {
		return
	}

	// If we have a state (update scenario), we might want to skip validation
	// if the plan is being destroyed
	if req.State.Raw.IsNull() && req.Plan.Raw.IsNull() {
		return
	}

	var plan credentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Count how many blocks are defined (not null and not unknown)
	blocksDefined := 0
	blockNames := []string{}

	if !plan.BasicAuth.IsNull() && !plan.BasicAuth.IsUnknown() {
		blocksDefined++
		blockNames = append(blockNames, "basic_auth")
	}
	if !plan.OAuth2.IsNull() && !plan.OAuth2.IsUnknown() {
		blocksDefined++
		blockNames = append(blockNames, "oauth2")
	}
	if !plan.HeaderAuth.IsNull() && !plan.HeaderAuth.IsUnknown() {
		blocksDefined++
		blockNames = append(blockNames, "header_auth")
	}

	// If all blocks are unknown, skip validation (might be during refresh)
	if plan.BasicAuth.IsUnknown() && plan.OAuth2.IsUnknown() && plan.HeaderAuth.IsUnknown() {
		return
	}

	// Validate exactly one block is provided
	if blocksDefined == 0 {
		resp.Diagnostics.AddError(
			"Missing Credential Block",
			"Exactly one credential block must be specified: basic_auth, oauth2, or header_auth",
		)
		return
	}
	if blocksDefined > 1 {
		resp.Diagnostics.AddError(
			"Multiple Credential Blocks",
			fmt.Sprintf("Exactly one credential block must be specified, but %d were found (%s). Please specify only one of: basic_auth, oauth2, or header_auth", blocksDefined, fmt.Sprintf("%v", blockNames)),
		)
		return
	}

	// Now validate that the selected block has all required attributes
	if !plan.BasicAuth.IsNull() && !plan.BasicAuth.IsUnknown() {
		var basicAuth basicAuthModel
		diags := plan.BasicAuth.As(ctx, &basicAuth, basetypes.ObjectAsOptions{})
		if !diags.HasError() {
			if basicAuth.Username.IsNull() || basicAuth.Username.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("basic_auth").AtName("username"),
					"Missing Required Attribute",
					"The username attribute is required when using the basic_auth block.",
				)
			}
			if basicAuth.Password.IsNull() || basicAuth.Password.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("basic_auth").AtName("password"),
					"Missing Required Attribute",
					"The password attribute is required when using the basic_auth block.",
				)
			}
		}
	}

	if !plan.OAuth2.IsNull() && !plan.OAuth2.IsUnknown() {
		var oauth2 oAuth2Model
		diags := plan.OAuth2.As(ctx, &oauth2, basetypes.ObjectAsOptions{})
		if !diags.HasError() {
			if oauth2.ClientId.IsNull() || oauth2.ClientId.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("oauth2").AtName("client_id"),
					"Missing Required Attribute",
					"The client_id attribute is required when using the oauth2 block.",
				)
			}
			if oauth2.ClientSecret.IsNull() || oauth2.ClientSecret.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("oauth2").AtName("client_secret"),
					"Missing Required Attribute",
					"The client_secret attribute is required when using the oauth2 block.",
				)
			}
			if oauth2.AccessTokenUrl.IsNull() || oauth2.AccessTokenUrl.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("oauth2").AtName("access_token_url"),
					"Missing Required Attribute",
					"The access_token_url attribute is required when using the oauth2 block.",
				)
			}
			if oauth2.AuthUrl.IsNull() || oauth2.AuthUrl.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("oauth2").AtName("auth_url"),
					"Missing Required Attribute",
					"The auth_url attribute is required when using the oauth2 block.",
				)
			}
			if oauth2.Scope.IsNull() || oauth2.Scope.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("oauth2").AtName("scope"),
					"Missing Required Attribute",
					"The scope attribute is required when using the oauth2 block.",
				)
			}
		}
	}

	if !plan.HeaderAuth.IsNull() && !plan.HeaderAuth.IsUnknown() {
		var headerAuth headerAuthModel
		diags := plan.HeaderAuth.As(ctx, &headerAuth, basetypes.ObjectAsOptions{})
		if !diags.HasError() {
			if headerAuth.Name.IsNull() || headerAuth.Name.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("header_auth").AtName("name"),
					"Missing Required Attribute",
					"The name attribute is required when using the header_auth block.",
				)
			}
			if headerAuth.Value.IsNull() || headerAuth.Value.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("header_auth").AtName("value"),
					"Missing Required Attribute",
					"The value attribute is required when using the header_auth block.",
				)
			}
		}
	}
}

// validateCredentialBlocks ensures exactly one credential block is defined.
func validateCredentialBlocks(ctx context.Context, model credentialResourceModel) (string, map[string]interface{}, error) {
	blocksDefined := 0
	var credentialType string
	var data map[string]interface{}

	if !model.BasicAuth.IsNull() && !model.BasicAuth.IsUnknown() {
		blocksDefined++
		//nolint:gosec // G101: This is a credential type identifier, not actual credentials
		credentialType = "httpBasicAuth"
		var basicAuth basicAuthModel
		diags := model.BasicAuth.As(ctx, &basicAuth, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return "", nil, fmt.Errorf("failed to parse basic_auth block: %v", diags)
		}
		data = map[string]interface{}{
			"user":     basicAuth.Username.ValueString(),
			"password": basicAuth.Password.ValueString(),
		}
	}

	if !model.OAuth2.IsNull() && !model.OAuth2.IsUnknown() {
		blocksDefined++
		//nolint:gosec // G101: This is a credential type identifier, not actual credentials
		credentialType = "oAuth2Api"
		var oauth2 oAuth2Model
		diags := model.OAuth2.As(ctx, &oauth2, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return "", nil, fmt.Errorf("failed to parse oauth2 block: %v", diags)
		}
		data = map[string]interface{}{
			"clientId":       oauth2.ClientId.ValueString(),
			"clientSecret":   oauth2.ClientSecret.ValueString(),
			"accessTokenUrl": oauth2.AccessTokenUrl.ValueString(),
			"authUrl":        oauth2.AuthUrl.ValueString(),
			"scope":          oauth2.Scope.ValueString(),
		}
		if !oauth2.AuthQueryParameters.IsNull() {
			data["authQueryParameters"] = oauth2.AuthQueryParameters.ValueString()
		} else {
			data["authQueryParameters"] = ""
		}
		if !oauth2.SendAdditionalBodyProperties.IsNull() {
			data["sendAdditionalBodyProperties"] = oauth2.SendAdditionalBodyProperties.ValueBool()
		} else {
			data["sendAdditionalBodyProperties"] = false
		}
		if !oauth2.AdditionalBodyProperties.IsNull() {
			data["additionalBodyProperties"] = oauth2.AdditionalBodyProperties.ValueString()
		} else {
			data["additionalBodyProperties"] = ""
		}
	}

	if !model.HeaderAuth.IsNull() && !model.HeaderAuth.IsUnknown() {
		blocksDefined++
		//nolint:gosec // G101: This is a credential type identifier, not actual credentials
		credentialType = "httpHeaderAuth"
		var headerAuth headerAuthModel
		diags := model.HeaderAuth.As(ctx, &headerAuth, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return "", nil, fmt.Errorf("failed to parse header_auth block: %v", diags)
		}
		data = map[string]interface{}{
			"name":  headerAuth.Name.ValueString(),
			"value": headerAuth.Value.ValueString(),
		}
	}

	if blocksDefined == 0 {
		return "", nil, fmt.Errorf("exactly one credential block must be specified (basic_auth, oauth2, or header_auth)")
	}
	if blocksDefined > 1 {
		return "", nil, fmt.Errorf("exactly one credential block must be specified, but %d were found", blocksDefined)
	}

	return credentialType, data, nil
}

// requiresReplaceListModifier is a plan modifier that marks the resource for replacement
// when the list attribute changes.
type requiresReplaceListModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m *requiresReplaceListModifier) Description(ctx context.Context) string {
	return "Requires replacement when nodes_access changes"
}

// MarkdownDescription returns a markdown formatted human-readable description of the plan modifier.
func (m *requiresReplaceListModifier) MarkdownDescription(ctx context.Context) string {
	return "Requires replacement when nodes_access changes"
}

// PlanModifyList implements the plan modification logic.
func (m *requiresReplaceListModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If the attribute is being removed or changed, require replacement
	if !req.StateValue.IsNull() && !req.PlanValue.IsNull() {
		// Check if values are different
		if !req.StateValue.Equal(req.PlanValue) {
			resp.RequiresReplace = true
		}
	} else if req.StateValue.IsNull() != req.PlanValue.IsNull() {
		// One is null and the other isn't - require replacement
		resp.RequiresReplace = true
	}
}

// requiresReplaceObjectModifier is a plan modifier that marks the resource for replacement
// when the object attribute changes.
type requiresReplaceObjectModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m *requiresReplaceObjectModifier) Description(ctx context.Context) string {
	return "Requires replacement when credential block changes"
}

// MarkdownDescription returns a markdown formatted human-readable description of the plan modifier.
func (m *requiresReplaceObjectModifier) MarkdownDescription(ctx context.Context) string {
	return "Requires replacement when credential block changes"
}

// PlanModifyObject implements the plan modification logic.
func (m *requiresReplaceObjectModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// If the attribute is being removed or changed, require replacement
	if !req.StateValue.IsNull() && !req.PlanValue.IsNull() {
		// Check if values are different
		if !req.StateValue.Equal(req.PlanValue) {
			resp.RequiresReplace = true
		}
	} else if req.StateValue.IsNull() != req.PlanValue.IsNull() {
		// One is null and the other isn't - require replacement
		resp.RequiresReplace = true
	}
}

// requiresReplaceBoolModifier is a plan modifier that marks the resource for replacement
// when the bool attribute changes.
type requiresReplaceBoolModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m *requiresReplaceBoolModifier) Description(ctx context.Context) string {
	return "Requires replacement when attribute changes"
}

// MarkdownDescription returns a markdown formatted human-readable description of the plan modifier.
func (m *requiresReplaceBoolModifier) MarkdownDescription(ctx context.Context) string {
	return "Requires replacement when attribute changes"
}

// PlanModifyBool implements the plan modification logic.
func (m *requiresReplaceBoolModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// If the attribute is being changed, require replacement
	if !req.StateValue.IsNull() && !req.PlanValue.IsNull() {
		if req.StateValue.ValueBool() != req.PlanValue.ValueBool() {
			resp.RequiresReplace = true
		}
	} else if req.StateValue.IsNull() != req.PlanValue.IsNull() {
		resp.RequiresReplace = true
	}
}
