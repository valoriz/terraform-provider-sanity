package dataset

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

// Ensure DatasetDataSource satisfies the datasource interfaces.
var (
	_ datasource.DataSource              = &DatasetDataSource{}
	_ datasource.DataSourceWithConfigure = &DatasetDataSource{}
)

// NewDatasetDataSource returns a new dataset data source instance.
func NewDatasetDataSource() datasource.DataSource {
	return &DatasetDataSource{}
}

// DatasetDataSource defines the data source implementation.
type DatasetDataSource struct {
	client *client.Client
}

// DatasetDataSourceModel maps the data source schema data to Go types.
type DatasetDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	ACLMode   types.String `tfsdk:"acl_mode"`
	CreatedAt types.String `tfsdk:"created_at"`
}

// Metadata returns the data source type name.
func (d *DatasetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

// Schema defines the schema for the data source.
func (d *DatasetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads information about an existing Sanity.io dataset within a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the format `<project_id>/<dataset_name>`.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the Sanity project to which the dataset belongs.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the dataset (e.g. `production`).",
				Required:    true,
			},
			"acl_mode": schema.StringAttribute{
				Description: "The ACL mode of the dataset (`public` or `private`).",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the dataset was created.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DatasetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the data source state with the latest dataset data.
func (d *DatasetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DatasetDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	name := state.Name.ValueString()

	tflog.Info(ctx, "Reading Sanity dataset data source", map[string]interface{}{
		"projectId": projectID,
		"name":      name,
	})

	dataset, err := d.client.GetDataset(ctx, projectID, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Sanity Dataset Data Source",
			fmt.Sprintf("Could not read dataset %q in project %q: %s", name, projectID, err.Error()),
		)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", projectID, name))
	state.ACLMode = types.StringValue(dataset.ACLMode)
	if !dataset.CreatedAt.IsZero() {
		state.CreatedAt = types.StringValue(dataset.CreatedAt.Format(time.RFC3339))
	} else {
		state.CreatedAt = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
