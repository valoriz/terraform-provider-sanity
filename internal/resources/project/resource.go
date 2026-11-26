package project

import (
	"context"
	"fmt"
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

// Ensure ProjectResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &ProjectResource{}
	_ resource.ResourceWithConfigure   = &ProjectResource{}
	_ resource.ResourceWithImportState = &ProjectResource{}
)

// NewProjectResource returns a new project resource instance.
func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// ProjectResource defines the resource implementation.
type ProjectResource struct {
	client *client.Client
}

// ProjectResourceModel maps the resource schema data to Go types.
type ProjectResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectName    types.String `tfsdk:"project_name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	StudioHost     types.String `tfsdk:"studio_host"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the resource.
func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sanity.io project. Projects are the top-level container for datasets, tokens, CORS origins, and webhooks.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Sanity project (e.g. `xyz12345`).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_name": schema.StringAttribute{
				Description: "The human-readable name of the Sanity project.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID to which this project belongs (e.g. `o-xyz12345`). Highly recommended so the project appears in your team's organization dashboard.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"studio_host": schema.StringAttribute{
				Description: "The custom studio subdomain hosted on sanity.studio (e.g. `my-app` maps to `my-app.sanity.studio`).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the project was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the project was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateProjectRequest{
		DisplayName: plan.ProjectName.ValueString(),
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		createReq.OrganizationID = plan.OrganizationID.ValueString()
	}

	tflog.Info(ctx, "Creating Sanity project", map[string]interface{}{
		"projectName": createReq.DisplayName,
	})

	project, err := r.client.CreateProject(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Sanity Project",
			fmt.Sprintf("Could not create project %q: %s", createReq.DisplayName, err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(project.ID)
	plan.ProjectName = types.StringValue(project.DisplayName)
	if project.OrganizationID != "" {
		plan.OrganizationID = types.StringValue(project.OrganizationID)
	} else {
		plan.OrganizationID = types.StringNull()
	}
	if project.StudioHost != nil && *project.StudioHost != "" {
		plan.StudioHost = types.StringValue(*project.StudioHost)
	} else {
		plan.StudioHost = types.StringNull()
	}
	plan.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(project.UpdatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ID.ValueString()
	tflog.Info(ctx, "Reading Sanity project", map[string]interface{}{
		"projectId": projectID,
	})

	project, err := r.client.GetProject(ctx, projectID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Sanity project not found, removing from state", map[string]interface{}{
				"projectId": projectID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Sanity Project",
			fmt.Sprintf("Could not read Sanity project %q: %s", projectID, err.Error()),
		)
		return
	}

	state.ProjectName = types.StringValue(project.DisplayName)
	if project.OrganizationID != "" {
		state.OrganizationID = types.StringValue(project.OrganizationID)
	}
	if project.StudioHost != nil && *project.StudioHost != "" {
		state.StudioHost = types.StringValue(*project.StudioHost)
	} else {
		state.StudioHost = types.StringNull()
	}
	state.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	state.UpdatedAt = types.StringValue(project.UpdatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ID.ValueString()
	projectName := plan.ProjectName.ValueString()

	updateReq := client.UpdateProjectRequest{
		DisplayName: &projectName,
	}

	tflog.Info(ctx, "Updating Sanity project", map[string]interface{}{
		"projectId": projectID,
	})

	project, err := r.client.UpdateProject(ctx, projectID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Sanity Project",
			fmt.Sprintf("Could not update Sanity project %q: %s", projectID, err.Error()),
		)
		return
	}

	plan.ProjectName = types.StringValue(project.DisplayName)
	if project.StudioHost != nil && *project.StudioHost != "" {
		plan.StudioHost = types.StringValue(*project.StudioHost)
	} else {
		plan.StudioHost = types.StringNull()
	}
	plan.UpdatedAt = types.StringValue(project.UpdatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ID.ValueString()
	tflog.Info(ctx, "Deleting Sanity project", map[string]interface{}{
		"projectId": projectID,
	})

	err := r.client.DeleteProject(ctx, projectID)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error Deleting Sanity Project",
			fmt.Sprintf("Could not delete Sanity project %q: %s", projectID, err.Error()),
		)
		return
	}
}

// ImportState imports an existing project into Terraform state using its project ID.
func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
