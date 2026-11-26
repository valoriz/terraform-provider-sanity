package cors_origin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/valoriz/terraform-provider-sanity/internal/client"
)

// Ensure CorsOriginResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &CorsOriginResource{}
	_ resource.ResourceWithConfigure   = &CorsOriginResource{}
	_ resource.ResourceWithImportState = &CorsOriginResource{}
)

// NewCorsOriginResource returns a new CORS origin resource instance.
func NewCorsOriginResource() resource.Resource {
	return &CorsOriginResource{}
}

// CorsOriginResource defines the resource implementation.
type CorsOriginResource struct {
	client *client.Client
}

// CorsOriginResourceModel maps the resource schema data to Go types.
type CorsOriginResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	Origin           types.String `tfsdk:"origin"`
	AllowCredentials types.Bool   `tfsdk:"allow_credentials"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *CorsOriginResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cors_origin"
}

// Schema defines the schema for the resource.
func (r *CorsOriginResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cross-Origin Resource Sharing (CORS) origin for a Sanity project. Allows web applications (like Sanity Studio or your frontend) to make authenticated or unauthenticated requests to the Sanity API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The internal numeric identifier assigned by Sanity (e.g. `3294326`).",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the Sanity project to configure CORS for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"origin": schema.StringAttribute{
				Description: "The URL origin including scheme and optional port (e.g. `https://my-app.com` or `http://localhost:3000`).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allow_credentials": schema.BoolAttribute{
				Description: "Whether to allow requests with credentials (cookies, authorization headers). Required for Sanity Studio login.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the CORS origin was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the CORS origin was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *CorsOriginResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new CORS origin.
func (r *CorsOriginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CorsOriginResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	originURL := plan.Origin.ValueString()
	allowCreds := plan.AllowCredentials.ValueBool()

	tflog.Info(ctx, "Creating Sanity CORS origin", map[string]interface{}{
		"projectId":        projectID,
		"origin":           originURL,
		"allowCredentials": allowCreds,
	})

	cors, err := r.client.CreateCorsOrigin(ctx, projectID, client.CreateCorsOriginRequest{
		Origin:           originURL,
		AllowCredentials: allowCreds,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Sanity CORS Origin",
			fmt.Sprintf("Could not create CORS origin %q for project %q: %s", originURL, projectID, err.Error()),
		)
		return
	}

	plan.ID = types.Int64Value(cors.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Origin = types.StringValue(cors.Origin)
	plan.AllowCredentials = types.BoolValue(cors.AllowCredentials)
	plan.CreatedAt = types.StringValue(cors.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(cors.UpdatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the state of the CORS origin.
func (r *CorsOriginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CorsOriginResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	corsID := state.ID.ValueInt64()

	tflog.Info(ctx, "Reading Sanity CORS origin", map[string]interface{}{
		"projectId": projectID,
		"corsId":    corsID,
	})

	cors, err := r.client.GetCorsOrigin(ctx, projectID, corsID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Sanity CORS origin not found, removing from state", map[string]interface{}{
				"projectId": projectID,
				"corsId":    corsID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Sanity CORS Origin",
			fmt.Sprintf("Could not read CORS origin %d for project %q: %s", corsID, projectID, err.Error()),
		)
		return
	}

	state.Origin = types.StringValue(cors.Origin)
	state.AllowCredentials = types.BoolValue(cors.AllowCredentials)
	state.CreatedAt = types.StringValue(cors.CreatedAt.Format(time.RFC3339))
	state.UpdatedAt = types.StringValue(cors.UpdatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update handles in-place updates. Since Sanity API does not have a CORS update endpoint,
// any attribute change is marked to require replacement.
func (r *CorsOriginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CorsOriginResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes a CORS origin.
func (r *CorsOriginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CorsOriginResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	corsID := state.ID.ValueInt64()

	tflog.Info(ctx, "Deleting Sanity CORS origin", map[string]interface{}{
		"projectId": projectID,
		"corsId":    corsID,
	})

	err := r.client.DeleteCorsOrigin(ctx, projectID, corsID)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error Deleting Sanity CORS Origin",
			fmt.Sprintf("Could not delete CORS origin %d in project %q: %s", corsID, projectID, err.Error()),
		)
		return
	}
}

// ImportState imports an existing CORS origin using the format `<project_id>/<cors_id>`.
func (r *CorsOriginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid CORS Origin Import ID",
			fmt.Sprintf("Expected import identifier in format '<project_id>/<cors_id>', got: %q", req.ID),
		)
		return
	}

	corsID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid CORS ID",
			fmt.Sprintf("CORS ID must be an integer, got: %q", parts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), corsID)...)
}
