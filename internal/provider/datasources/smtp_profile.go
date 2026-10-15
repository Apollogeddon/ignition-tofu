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
var _ datasource.DataSource = &SMTPProfileDataSource{}

func NewSMTPProfileDataSource() datasource.DataSource {
	return &SMTPProfileDataSource{}
}

// SMTPProfileDataSource defines the data source implementation.
type SMTPProfileDataSource struct {
	client client.IgnitionClient
}

// SMTPProfileDataSourceModel describes the data source data model.
type SMTPProfileDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Hostname        types.String `tfsdk:"hostname"`
	Port            types.Int64  `tfsdk:"port"`
	UseSslPort      types.Bool   `tfsdk:"use_ssl_port"`
	StartTlsEnabled types.Bool   `tfsdk:"start_tls_enabled"`
	Username        types.String `tfsdk:"username"`
}

func (d *SMTPProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp_profile"
}

func (d *SMTPProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an SMTP Email Profile from Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the SMTP profile.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the SMTP profile.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the SMTP profile is enabled.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the SMTP server.",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port SMTP service is running on.",
				Computed:    true,
			},
			"use_ssl_port": schema.BoolAttribute{
				Description: "Connect using dedicated SSL/TLS port.",
				Computed:    true,
			},
			"start_tls_enabled": schema.BoolAttribute{
				Description: "Connect using STARTTLS.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username for logging into the email server.",
				Computed:    true,
			},
		},
	}
}

func (d *SMTPProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SMTPProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		base.AddUnconfiguredError(&resp.Diagnostics)
		return
	}
	var data SMTPProfileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetSMTPProfile(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading SMTP profile", err.Error())
		return
	}

	data.Id = types.StringValue(res.Name)
	if res.Enabled != nil {
		data.Enabled = types.BoolValue(*res.Enabled)
	} else {
		data.Enabled = types.BoolValue(true)
	}

	if res.Description != "" {
		data.Description = types.StringValue(res.Description)
	} else {
		data.Description = types.StringNull()
	}

	if res.Config.Settings.Settings != nil {
		data.Hostname = types.StringValue(res.Config.Settings.Settings.Hostname)
		data.Port = types.Int64Value(int64(res.Config.Settings.Settings.Port))
		data.UseSslPort = types.BoolValue(res.Config.Settings.Settings.UseSslPort)
		data.StartTlsEnabled = types.BoolValue(res.Config.Settings.Settings.StartTlsEnabled)
		if res.Config.Settings.Settings.Username != "" {
			data.Username = types.StringValue(res.Config.Settings.Settings.Username)
		} else {
			data.Username = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
