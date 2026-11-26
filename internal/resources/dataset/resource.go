package dataset

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/valoriz/terraform-provider-sanity/internal/client"
)

// Ensure DatasetResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &DatasetResource{}
	_ resource.ResourceWithConfigure   = &DatasetResource{}
	_ resource.ResourceWithImportState = &DatasetResource{}
)

// NewDatasetResource returns a new dataset resource instance.
func NewDatasetResource() resource.Resource {
	return &DatasetResource{}
}

// DatasetResource defines the resource implementation.
type DatasetResource struct {
	client *client.Client
}

// DatasetResourceModel maps the resource schema data to Go types.
type DatasetResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	ACLMode   types.String `tfsdk:"acl_mode"`
	CreatedAt types.String `tfsdk:"created_at"`
}

// Metadata returns the resource type name.
func (r *DatasetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

// Schema defines the schema for the resource.
func (r *DatasetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sanity.io dataset within a project. Datasets hold documents and assets.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the format `<project_id>/<dataset_name>`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the Sanity project to which this dataset belongs.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the dataset (e.g. `production` or `staging`). Dataset names cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"acl_mode": schema.StringAttribute{
				Description: "Access Control List mode: `public` or `private`. Note: `private` requires a paid Growth or Enterprise plan.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the dataset was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *DatasetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = apiClient
}

// Create creates the dataset via PUT /projects/{projectId}/datasets/{name}.
func (r *DatasetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatasetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	name := plan.Name.ValueString()
	aclMode := "public"
	if !plan.ACLMode.IsNull() && plan.ACLMode.ValueString() != "" {
		aclMode = plan.ACLMode.ValueString()
	}

	tflog.Info(ctx, "Creating Sanity dataset", map[string]interface{}{
		"projectId": projectID,
		"name":      name,
		"aclMode":   aclMode,
	})

	dataset, err := r.client.CreateDataset(ctx, projectID, name, client.CreateDatasetRequest{
		ACLMode: aclMode,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Sanity Dataset",
			fmt.Sprintf("Could not create dataset %q in project %q: %s", name, projectID, err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", projectID, name))
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(name)
	plan.ACLMode = types.StringValue(dataset.ACLMode)
	if !dataset.CreatedAt.IsZero() {
		plan.CreatedAt = types.StringValue(dataset.CreatedAt.Format(time.RFC3339))
	} else {
		plan.CreatedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest dataset data.
func (r *DatasetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatasetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	name := state.Name.ValueString()

	tflog.Info(ctx, "Reading Sanity dataset", map[string]interface{}{
		"projectId": projectID,
		"name":      name,
	})

	dataset, err := r.client.GetDataset(ctx, projectID, name)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Sanity dataset not found, removing from state", map[string]interface{}{
				"projectId": projectID,
				"name":      name,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Sanity Dataset",
			fmt.Sprintf("Could not read dataset %q in project %q: %s", name, projectID, err.Error()),
		)
		return
	}

	state.ACLMode = types.StringValue(dataset.ACLMode)
	if !dataset.CreatedAt.IsZero() {
		state.CreatedAt = types.StringValue(dataset.CreatedAt.Format(time.RFC3339))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update handles updates. Since dataset name, project ID, and ACL mode all require replacement,
// this method is a no-op that preserves state.
func (r *DatasetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DatasetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes the dataset via DELETE /projects/{projectId}/datasets/{name}.
func (r *DatasetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatasetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	name := state.Name.ValueString()

	tflog.Info(ctx, "Deleting Sanity dataset", map[string]interface{}{
		"projectId": projectID,
		"name":      name,
	})

	err := r.client.DeleteDataset(ctx, projectID, name)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error Deleting Sanity Dataset",
			fmt.Sprintf("Could not delete dataset %q in project %q: %s", name, projectID, err.Error()),
		)
		return
	}
}

// ImportState imports an existing dataset using the format `<project_id>/<dataset_name>`.
func (r *DatasetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Dataset Import ID",
			fmt.Sprintf("Expected import identifier in format '<project_id>/<dataset_name>', got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}
