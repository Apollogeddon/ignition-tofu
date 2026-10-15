package datasources

import (
	"context"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DatabaseConnectionDataSource{}

func NewDatabaseConnectionDataSource() datasource.DataSource {
	return &DatabaseConnectionDataSource{}
}

// DatabaseConnectionDataSource defines the data source implementation.
type DatabaseConnectionDataSource struct {
	client client.IgnitionClient
}

// DatabaseConnectionDataSourceModel describes the data source data model.
type DatabaseConnectionDataSourceModel struct {
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	ConnectURL types.String `tfsdk:"connect_url"`
	Username   types.String `tfsdk:"username"`
}

func (d *DatabaseConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_connection"
}

func (d *DatabaseConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Database Connection from Ignition.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the database connection.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the database connection.",
				Computed:    true,
			},
			"connect_url": schema.StringAttribute{
				Description: "The JDBC connection URL.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username for the database connection.",
				Computed:    true,
			},
		},
	}
}

func (d *DatabaseConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.IgnitionClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.IgnitionClient, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DatabaseConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		base.AddUnconfiguredError(&resp.Diagnostics)
		return
	}
	var data DatabaseConnectionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	db, err := d.client.GetDatabaseConnection(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading database connection",
			"Could not read database connection "+data.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	data.Type = types.StringValue(db.Config.Driver)
	data.ConnectURL = types.StringValue(db.Config.ConnectURL)
	if db.Config.Username != "" {
		data.Username = types.StringValue(db.Config.Username)
	} else {
		data.Username = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
