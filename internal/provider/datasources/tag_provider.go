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
var _ datasource.DataSource = &TagProviderDataSource{}

func NewTagProviderDataSource() datasource.DataSource {
	return &TagProviderDataSource{}
}

// TagProviderDataSource defines the data source implementation.
type TagProviderDataSource struct {
	client client.IgnitionClient
}

// TagProviderDataSourceModel describes the data source data model.
type TagProviderDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (d *TagProviderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag_provider"
}

func (d *TagProviderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Tag Provider from Ignition.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the tag provider.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the tag provider.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the tag provider.",
				Computed:    true,
			},
		},
	}
}

func (d *TagProviderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TagProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		base.AddUnconfiguredError(&resp.Diagnostics)
		return
	}
	var data TagProviderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetTagProvider(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading tag provider", err.Error())
		return
	}

	data.Type = types.StringValue(res.Config.Profile.Type)
	if res.Config.Description != "" {
		data.Description = types.StringValue(res.Config.Description)
	} else {
		data.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
