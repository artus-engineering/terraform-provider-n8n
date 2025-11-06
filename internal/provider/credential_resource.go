package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/artus-engineering/terraform-provider-n8n/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &credentialResource{}
	_ resource.ResourceWithConfigure   = &credentialResource{}
	_ resource.ResourceWithImportState = &credentialResource{}
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
	Type        types.String `tfsdk:"type"`
	Data        types.String `tfsdk:"data"`
	NodesAccess types.List   `tfsdk:"nodes_access"`
}

// Metadata returns the resource type name.
func (r *credentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

// Schema defines the schema for the resource.
func (r *credentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a credential in n8n. Credentials are used to authenticate with external services.",
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
			"type": schema.StringAttribute{
				Description: "The type of credential (e.g., 'httpBasicAuth', 'httpHeaderAuth', 'oAuth2Api', etc.).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				Description: "JSON string containing the credential data. This is sensitive information and will be stored in state.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nodes_access": schema.ListAttribute{
				Description: "List of node types that can access this credential. Each item should be a string representing the node type.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					// Require replacement when nodes_access changes
					&requiresReplaceListModifier{},
				},
			},
		},
	}
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

	tflog.Info(ctx, "Creating credential", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	})

	// Parse the data JSON string
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(plan.Data.ValueString()), &data); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Data Format",
			fmt.Sprintf("The 'data' field must be a valid JSON string. Error: %s", err.Error()),
		)
		return
	}

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
		Type:        plan.Type.ValueString(),
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
	plan.Type = types.StringValue(createdCredential.Type)

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
	} else if !plan.NodesAccess.IsNull() {
		// Keep the original value if it was set
		plan.NodesAccess = plan.NodesAccess
	}

	// Note: We don't update the data field from the response because n8n API
	// doesn't return sensitive credential data for security reasons.
	// The data field remains as provided by the user.

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
	state.Type = types.StringValue(credential.Type)

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
func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan credentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating credential", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	// Parse the data JSON string
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(plan.Data.ValueString()), &data); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Data Format",
			fmt.Sprintf("The 'data' field must be a valid JSON string. Error: %s", err.Error()),
		)
		return
	}

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
		Type:        plan.Type.ValueString(),
		Data:        data,
		NodesAccess: nodesAccess,
	}

	// Update credential by deleting and recreating (n8n API doesn't support PUT/PATCH)
	// Note: This will result in a new credential ID
	tflog.Info(ctx, "Updating credential via delete-and-recreate", map[string]interface{}{
		"old_id": plan.ID.ValueString(),
		"name":   plan.Name.ValueString(),
	})

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
	plan.Type = types.StringValue(updatedCredential.Type)

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
	} else if !plan.NodesAccess.IsNull() {
		plan.NodesAccess = plan.NodesAccess
	}

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
