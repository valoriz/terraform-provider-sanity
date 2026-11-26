package webhook

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/valoriz/terraform-provider-sanity/internal/client"
)

// Ensure WebhookResource satisfies the resource interfaces.
var (
	_ resource.Resource                = &WebhookResource{}
	_ resource.ResourceWithConfigure   = &WebhookResource{}
	_ resource.ResourceWithImportState = &WebhookResource{}
)

// NewWebhookResource returns a new webhook resource instance.
func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

// WebhookResource defines the resource implementation.
type WebhookResource struct {
	client *client.Client
}

// WebhookResourceModel maps the resource schema data to Go types.
type WebhookResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	Name               types.String `tfsdk:"name"`
	URL                types.String `tfsdk:"url"`
	Dataset            types.String `tfsdk:"dataset"`
	APIVersion         types.String `tfsdk:"api_version"`
	HTTPMethod         types.String `tfsdk:"http_method"`
	Secret             types.String `tfsdk:"secret"`
	Filter             types.String `tfsdk:"filter"`
	Projection         types.String `tfsdk:"projection"`
	Headers            types.Map    `tfsdk:"headers"`
	IncludeDrafts      types.Bool   `tfsdk:"include_drafts"`
	IncludeAllVersions types.Bool   `tfsdk:"include_all_versions"`
	IsDisabledByUser   types.Bool   `tfsdk:"is_disabled_by_user"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

// Metadata returns the resource type name.
func (r *WebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

// Schema defines the schema for the resource.
func (r *WebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sanity.io document webhook. Webhooks trigger HTTPS POST requests when documents are created, updated, or deleted.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique string identifier of the webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the Sanity project to configure the webhook for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The human-readable name of the webhook.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The HTTPS URL destination that receives webhook payloads.",
				Required:    true,
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset to monitor for events (e.g. `production` or `*` for all).",
				Required:    true,
			},
			"api_version": schema.StringAttribute{
				Description: "The Sanity API version used to execute the filter and projection queries (e.g. `v2024-01-01`). Required by Sanity.",
				Required:    true,
			},
			"http_method": schema.StringAttribute{
				Description: "HTTP method used for webhook delivery: `POST`, `PUT`, `PATCH`, `DELETE`, `GET`. Defaults to `POST`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("POST"),
			},
			"secret": schema.StringAttribute{
				Description: "Shared secret used to sign the webhook payload using HMAC-SHA256 (`sanity-webhook-signature` header).",
				Optional:    true,
				Sensitive:   true,
			},
			"filter": schema.StringAttribute{
				Description: "GROQ filter expression to filter documents (e.g. `_type == \"post\" && !(_id in path(\"drafts.**\"))`).",
				Optional:    true,
			},
			"projection": schema.StringAttribute{
				Description: "GROQ projection expression to shape the payload JSON body.",
				Optional:    true,
			},
			"headers": schema.MapAttribute{
				Description: "Custom HTTP headers sent with the webhook delivery requests.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"include_drafts": schema.BoolAttribute{
				Description: "Whether to trigger the webhook on draft document mutations.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"include_all_versions": schema.BoolAttribute{
				Description: "Whether to trigger the webhook across all document releases.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_disabled_by_user": schema.BoolAttribute{
				Description: "Temporarily disables webhook execution without deleting the configuration.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the webhook was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *WebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create registers a new webhook.
func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()

	createReq := client.CreateWebhookRequest{
		Type:               "document",
		Name:               plan.Name.ValueString(),
		URL:                plan.URL.ValueString(),
		Dataset:            plan.Dataset.ValueString(),
		APIVersion:         plan.APIVersion.ValueString(),
		HTTPMethod:         plan.HTTPMethod.ValueString(),
		IncludeDrafts:      plan.IncludeDrafts.ValueBool(),
		IncludeAllVersions: plan.IncludeAllVersions.ValueBool(),
		IsDisabledByUser:   plan.IsDisabledByUser.ValueBool(),
	}

	if !plan.Secret.IsNull() {
		createReq.Secret = plan.Secret.ValueString()
	}
	if !plan.Filter.IsNull() {
		filterStr := plan.Filter.ValueString()
		createReq.Filter = &filterStr
	}
	if !plan.Projection.IsNull() {
		projStr := plan.Projection.ValueString()
		createReq.Projection = &projStr
	}
	if !plan.Headers.IsNull() {
		headersMap := make(map[string]string)
		diags := plan.Headers.ElementsAs(ctx, &headersMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Headers = headersMap
	}

	tflog.Info(ctx, "Creating Sanity webhook", map[string]interface{}{
		"projectId": projectID,
		"name":      createReq.Name,
		"url":       createReq.URL,
	})

	webhook, err := r.client.CreateWebhook(ctx, projectID, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Sanity Webhook",
			fmt.Sprintf("Could not create webhook %q for project %q: %s", createReq.Name, projectID, err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(webhook.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(webhook.Name)
	plan.URL = types.StringValue(webhook.URL)
	plan.Dataset = types.StringValue(webhook.Dataset)
	plan.APIVersion = types.StringValue(webhook.APIVersion)
	plan.HTTPMethod = types.StringValue(webhook.HTTPMethod)
	if webhook.Secret != "" {
		plan.Secret = types.StringValue(webhook.Secret)
	}
	plan.IncludeDrafts = types.BoolValue(webhook.IncludeDrafts)
	plan.IncludeAllVersions = types.BoolValue(webhook.IncludeAllVersions)
	plan.IsDisabledByUser = types.BoolValue(webhook.IsDisabledByUser)
	plan.CreatedAt = types.StringValue(webhook.CreatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the state of the webhook.
func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	webhookID := state.ID.ValueString()

	tflog.Info(ctx, "Reading Sanity webhook", map[string]interface{}{
		"projectId": projectID,
		"webhookId": webhookID,
	})

	webhook, err := r.client.GetWebhook(ctx, projectID, webhookID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Sanity webhook not found, removing from state", map[string]interface{}{
				"projectId": projectID,
				"webhookId": webhookID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Sanity Webhook",
			fmt.Sprintf("Could not read webhook %q for project %q: %s", webhookID, projectID, err.Error()),
		)
		return
	}

	state.Name = types.StringValue(webhook.Name)
	state.URL = types.StringValue(webhook.URL)
	state.Dataset = types.StringValue(webhook.Dataset)
	state.APIVersion = types.StringValue(webhook.APIVersion)
	state.HTTPMethod = types.StringValue(webhook.HTTPMethod)
	if webhook.Secret != "" {
		state.Secret = types.StringValue(webhook.Secret)
	}
	if webhook.Filter != nil {
		state.Filter = types.StringValue(*webhook.Filter)
	}
	if webhook.Projection != nil {
		state.Projection = types.StringValue(*webhook.Projection)
	}
	state.IncludeDrafts = types.BoolValue(webhook.IncludeDrafts)
	state.IncludeAllVersions = types.BoolValue(webhook.IncludeAllVersions)
	state.IsDisabledByUser = types.BoolValue(webhook.IsDisabledByUser)
	state.CreatedAt = types.StringValue(webhook.CreatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates an existing webhook in-place using PATCH.
func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	webhookID := plan.ID.ValueString()

	name := plan.Name.ValueString()
	urlStr := plan.URL.ValueString()
	dataset := plan.Dataset.ValueString()
	apiVer := plan.APIVersion.ValueString()
	method := plan.HTTPMethod.ValueString()
	disabled := plan.IsDisabledByUser.ValueBool()
	includeDrafts := plan.IncludeDrafts.ValueBool()
	includeAllVersions := plan.IncludeAllVersions.ValueBool()

	updateReq := client.UpdateWebhookRequest{
		Name:               &name,
		URL:                &urlStr,
		Dataset:            &dataset,
		APIVersion:         &apiVer,
		HTTPMethod:         &method,
		IncludeDrafts:      &includeDrafts,
		IncludeAllVersions: &includeAllVersions,
		IsDisabledByUser:   &disabled,
	}

	if !plan.Secret.IsNull() {
		sec := plan.Secret.ValueString()
		updateReq.Secret = &sec
	}
	if !plan.Filter.IsNull() {
		f := plan.Filter.ValueString()
		updateReq.Filter = &f
	}
	if !plan.Projection.IsNull() {
		p := plan.Projection.ValueString()
		updateReq.Projection = &p
	}

	tflog.Info(ctx, "Updating Sanity webhook", map[string]interface{}{
		"projectId": projectID,
		"webhookId": webhookID,
	})

	webhook, err := r.client.UpdateWebhook(ctx, projectID, webhookID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Sanity Webhook",
			fmt.Sprintf("Could not update webhook %q in project %q: %s", webhookID, projectID, err.Error()),
		)
		return
	}

	plan.Name = types.StringValue(webhook.Name)
	plan.URL = types.StringValue(webhook.URL)
	plan.Dataset = types.StringValue(webhook.Dataset)
	plan.APIVersion = types.StringValue(webhook.APIVersion)
	plan.HTTPMethod = types.StringValue(webhook.HTTPMethod)
	plan.IncludeDrafts = types.BoolValue(webhook.IncludeDrafts)
	plan.IncludeAllVersions = types.BoolValue(webhook.IncludeAllVersions)
	plan.IsDisabledByUser = types.BoolValue(webhook.IsDisabledByUser)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes a webhook.
func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	webhookID := state.ID.ValueString()

	tflog.Info(ctx, "Deleting Sanity webhook", map[string]interface{}{
		"projectId": projectID,
		"webhookId": webhookID,
	})

	err := r.client.DeleteWebhook(ctx, projectID, webhookID)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error Deleting Sanity Webhook",
			fmt.Sprintf("Could not delete webhook %q in project %q: %s", webhookID, projectID, err.Error()),
		)
		return
	}
}

// ImportState imports an existing webhook using `<project_id>/<webhook_id>`.
func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Webhook Import ID",
			fmt.Sprintf("Expected import identifier in format '<project_id>/<webhook_id>', got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
