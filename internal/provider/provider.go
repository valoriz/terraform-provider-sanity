package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/valoriz/terraform-provider-sanity/internal/client"
	datasetds "github.com/valoriz/terraform-provider-sanity/internal/datasources/dataset"
	projectds "github.com/valoriz/terraform-provider-sanity/internal/datasources/project"
	"github.com/valoriz/terraform-provider-sanity/internal/resources/cors_origin"
	"github.com/valoriz/terraform-provider-sanity/internal/resources/dataset"
	"github.com/valoriz/terraform-provider-sanity/internal/resources/invitation"
	"github.com/valoriz/terraform-provider-sanity/internal/resources/project"
	"github.com/valoriz/terraform-provider-sanity/internal/resources/webhook"
)

// Ensure SanityProvider satisfies the provider.Provider interface.
var _ provider.Provider = &SanityProvider{}

// SanityProvider defines the provider implementation.
type SanityProvider struct {
	version string
}

// SanityProviderModel maps provider-level schema configuration to Go types.
type SanityProviderModel struct {
	APIToken   types.String `tfsdk:"api_token"`
	APIVersion types.String `tfsdk:"api_version"`
}

// New returns a new provider instance constructor.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SanityProvider{
			version: version,
		}
	}
}

// Metadata returns provider metadata like name.
func (p *SanityProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sanity"
	resp.Version = p.version
}

// Schema defines the provider-level configuration schema.
func (p *SanityProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Sanity.io platform resources (projects, datasets, CORS origins, webhooks).",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "Sanity API token. Can also be set via the SANITY_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_version": schema.StringAttribute{
				Description: "Sanity API version to use (e.g. v2024-01-01). Defaults to v2024-01-01. Can also be set via SANITY_API_VERSION.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares the Sanity API client for data sources and resources.
func (p *SanityProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Sanity provider client")

	var config SanityProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 1. Resolve API Token
	token := os.Getenv("SANITY_TOKEN")
	if !config.APIToken.IsNull() && config.APIToken.ValueString() != "" {
		token = config.APIToken.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Sanity API Token",
			"The Sanity API token must be provided via the `api_token` argument in the provider block or the SANITY_TOKEN environment variable.",
		)
		return
	}

	// 2. Resolve API Version
	apiVersion := client.DefaultAPIVersion
	if envVersion := os.Getenv("SANITY_API_VERSION"); envVersion != "" {
		apiVersion = envVersion
	}
	if !config.APIVersion.IsNull() && config.APIVersion.ValueString() != "" {
		apiVersion = config.APIVersion.ValueString()
	}

	// 3. Initialize API client
	apiClient, err := client.NewClient(token, client.WithAPIVersion(apiVersion))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Initialize Sanity Client",
			"Unable to create Sanity API client: "+err.Error(),
		)
		return
	}

	// Make the client available to all resources and data sources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient

	tflog.Info(ctx, "Sanity provider client configured successfully")
}

// Resources defines the resources implemented in the provider.
func (p *SanityProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		project.NewProjectResource,
		dataset.NewDatasetResource,
		cors_origin.NewCorsOriginResource,
		webhook.NewWebhookResource,
		invitation.NewInvitationResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *SanityProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		projectds.NewProjectDataSource,
		datasetds.NewDatasetDataSource,
	}
}
