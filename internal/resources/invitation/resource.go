package invitation

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

// Ensure InvitationResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &InvitationResource{}
	_ resource.ResourceWithConfigure   = &InvitationResource{}
	_ resource.ResourceWithImportState = &InvitationResource{}
)

// NewInvitationResource returns a new invitation resource instance.
func NewInvitationResource() resource.Resource {
	return &InvitationResource{}
}

// InvitationResource defines the resource implementation.
type InvitationResource struct {
	client *client.Client
}

// InvitationResourceModel maps the resource schema data to Go types.
type InvitationResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	Email      types.String `tfsdk:"email"`
	Role       types.String `tfsdk:"role"`
	IsAccepted types.Bool   `tfsdk:"is_accepted"`
	IsRevoked  types.Bool   `tfsdk:"is_revoked"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

// Metadata returns the resource type name.
func (r *InvitationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_invitation"
}

// Schema defines the schema for the resource.
func (r *InvitationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a project member invitation for a Sanity.io project. Sends an email invitation granting the recipient access to the project with the specified role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the invitation (e.g. `vXJYs4Is`).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the Sanity project to invite the member to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Description: "The email address of the person to invite.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Description: "The role to assign to the invited user (e.g. `viewer`, `editor`, `developer`, `administrator`).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_accepted": schema.BoolAttribute{
				Description: "Whether the user has accepted the invitation.",
				Computed:    true,
			},
			"is_revoked": schema.BoolAttribute{
				Description: "Whether the invitation has been revoked.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the invitation was sent.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *InvitationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create sends a new invitation.
func (r *InvitationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InvitationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	email := plan.Email.ValueString()
	role := plan.Role.ValueString()

	tflog.Info(ctx, "Inviting member to Sanity project", map[string]interface{}{
		"projectId": projectID,
		"email":     email,
		"role":      role,
	})

	inv, err := r.client.CreateInvitation(ctx, projectID, client.CreateInvitationRequest{
		Email: email,
		Role:  role,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Inviting Sanity Project Member",
			fmt.Sprintf("Could not invite member %q to project %q: %s", email, projectID, err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(inv.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Email = types.StringValue(inv.Email)
	plan.Role = types.StringValue(role)
	plan.IsAccepted = types.BoolValue(inv.IsAccepted)
	plan.IsRevoked = types.BoolValue(inv.IsRevoked)
	if !inv.CreatedAt.IsZero() {
		plan.CreatedAt = types.StringValue(inv.CreatedAt.Format(time.RFC3339))
	} else {
		plan.CreatedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the state of the invitation.
func (r *InvitationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InvitationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	invID := state.ID.ValueString()

	tflog.Info(ctx, "Reading Sanity project invitation", map[string]interface{}{
		"projectId":    projectID,
		"invitationId": invID,
	})

	inv, err := r.client.GetInvitation(ctx, projectID, invID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Sanity project invitation not found, removing from state", map[string]interface{}{
				"projectId":    projectID,
				"invitationId": invID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Sanity Project Invitation",
			fmt.Sprintf("Could not read invitation %q for project %q: %s", invID, projectID, err.Error()),
		)
		return
	}

	if inv.IsRevoked {
		tflog.Warn(ctx, "Sanity project invitation was revoked, removing from state", map[string]interface{}{
			"projectId":    projectID,
			"invitationId": invID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	state.Email = types.StringValue(inv.Email)
	if len(inv.Roles) > 0 {
		state.Role = types.StringValue(inv.Roles[0].Name)
	}
	state.IsAccepted = types.BoolValue(inv.IsAccepted)
	state.IsRevoked = types.BoolValue(inv.IsRevoked)
	if !inv.CreatedAt.IsZero() {
		state.CreatedAt = types.StringValue(inv.CreatedAt.Format(time.RFC3339))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update handles in-place updates. Changing email, project, or role requires replacement.
func (r *InvitationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InvitationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete revokes the invitation.
func (r *InvitationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InvitationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	invID := state.ID.ValueString()

	tflog.Info(ctx, "Revoking Sanity project invitation", map[string]interface{}{
		"projectId":    projectID,
		"invitationId": invID,
	})

	err := r.client.DeleteInvitation(ctx, projectID, invID)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error Revoking Sanity Project Invitation",
			fmt.Sprintf("Could not revoke invitation %q in project %q: %s", invID, projectID, err.Error()),
		)
		return
	}
}

// ImportState imports an existing invitation using `<project_id>/<invitation_id>`.
func (r *InvitationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Invitation Import ID",
			fmt.Sprintf("Expected import identifier in format '<project_id>/<invitation_id>', got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
