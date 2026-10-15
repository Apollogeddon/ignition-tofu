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
var _ datasource.DataSource = &UserSourceDataSource{}

func NewUserSourceDataSource() datasource.DataSource {
	return &UserSourceDataSource{}
}

// UserSourceDataSource defines the data source implementation.
type UserSourceDataSource struct {
	client client.IgnitionClient
}

// UserSourceDataSourceModel describes the data source data model.
type UserSourceDataSourceModel struct {
	Name               types.String `tfsdk:"name"`
	Type               types.String `tfsdk:"type"`
	FailoverProfile    types.String `tfsdk:"failover_profile"`
	FailoverMode       types.String `tfsdk:"failover_mode"`
	ScheduleRestricted types.Bool   `tfsdk:"schedule_restricted"`
}

func (d *UserSourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_source"
}

func (d *UserSourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a User Source from Ignition.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the user source.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the user source profile.",
				Computed:    true,
			},
			"failover_profile": schema.StringAttribute{
				Description: "The failover profile name.",
				Computed:    true,
			},
			"failover_mode": schema.StringAttribute{
				Description: "The failover mode.",
				Computed:    true,
			},
			"schedule_restricted": schema.BoolAttribute{
				Description: "Whether the user source is schedule restricted.",
				Computed:    true,
			},
		},
	}
}

func (d *UserSourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserSourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		base.AddUnconfiguredError(&resp.Diagnostics)
		return
	}
	var data UserSourceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetUserSource(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user source", err.Error())
		return
	}

	data.Type = types.StringValue(res.Config.Profile.Type)
	if res.Config.Profile.FailoverProfile != "" {
		data.FailoverProfile = types.StringValue(res.Config.Profile.FailoverProfile)
	} else {
		data.FailoverProfile = types.StringNull()
	}
	if res.Config.Profile.FailoverMode != "" {
		data.FailoverMode = types.StringValue(res.Config.Profile.FailoverMode)
	} else {
		data.FailoverMode = types.StringNull()
	}
	data.ScheduleRestricted = types.BoolValue(base.Deref(res.Config.Profile.ScheduleRestricted))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
