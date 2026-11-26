package project

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/valoriz/terraform-provider-sanity/internal/client"
)

// Ensure ProjectDataSource satisfies the datasource interfaces.
var (
	_ datasource.DataSource              = &ProjectDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectDataSource{}
)

// NewProjectDataSource returns a new project data source instance.
func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource defines the data source implementation.
type ProjectDataSource struct {
	client *client.Client
}

// ProjectDataSourceModel maps the data source schema data to Go types.
type ProjectDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectName    types.String `tfsdk:"project_name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	StudioHost     types.String `tfsdk:"studio_host"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the data source.
func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads information about an existing Sanity.io project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Sanity project (e.g. `xyz12345`).",
				Required:    true,
			},
			"project_name": schema.StringAttribute{
				Description: "The human-readable name of the Sanity project.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID to which this project belongs.",
				Computed:    true,
			},
			"studio_host": schema.StringAttribute{
				Description: "The custom studio subdomain hosted on sanity.studio.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the project was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the project was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = apiClient
}

// Read refreshes the data source state with the latest project data.
func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ProjectDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ID.ValueString()
	tflog.Info(ctx, "Reading Sanity project data source", map[string]interface{}{
		"projectId": projectID,
	})

	project, err := d.client.GetProject(ctx, projectID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Sanity Project Data Source",
			fmt.Sprintf("Could not read project %q: %s", projectID, err.Error()),
		)
		return
	}

	state.ProjectName = types.StringValue(project.DisplayName)
	if project.OrganizationID != "" {
		state.OrganizationID = types.StringValue(project.OrganizationID)
	} else {
		state.OrganizationID = types.StringNull()
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
