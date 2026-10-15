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
var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource defines the data source implementation.
type ProjectDataSource struct {
	client client.IgnitionClient
}

// ProjectDataSourceModel describes the data source data model.
type ProjectDataSourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Title            types.String `tfsdk:"title"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Parent           types.String `tfsdk:"parent"`
	Inheritable      types.Bool   `tfsdk:"inheritable"`
	DefaultDB        types.String `tfsdk:"default_db"`
	TagProvider      types.String `tfsdk:"tag_provider"`
	UserSource       types.String `tfsdk:"user_source"`
	IdentityProvider types.String `tfsdk:"identity_provider"`
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an Ignition Project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the project.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the project.",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "The title of the project.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the project is enabled.",
				Computed:    true,
			},
			"parent": schema.StringAttribute{
				Description: "The parent project name, if any.",
				Computed:    true,
			},
			"inheritable": schema.BoolAttribute{
				Description: "Whether the project is inheritable.",
				Computed:    true,
			},
			"default_db": schema.StringAttribute{
				Description: "The default database connection for the project.",
				Computed:    true,
			},
			"tag_provider": schema.StringAttribute{
				Description: "The default tag provider for the project.",
				Computed:    true,
			},
			"user_source": schema.StringAttribute{
				Description: "The default user source for the project.",
				Computed:    true,
			},
			"identity_provider": schema.StringAttribute{
				Description: "The default identity provider for the project.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetProject(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	data.Description = types.StringValue(res.Description)
	data.Id = types.StringValue(data.Name.ValueString())
	if res.Title != "" {
		data.Title = types.StringValue(res.Title)
	} else {
		data.Title = types.StringNull()
	}
	data.Enabled = types.BoolValue(base.Deref(res.Enabled))
	if res.Parent != "" {
		data.Parent = types.StringValue(res.Parent)
	} else {
		data.Parent = types.StringNull()
	}
	data.Inheritable = types.BoolValue(base.Deref(res.Inheritable))
	if res.DefaultDB != "" {
		data.DefaultDB = types.StringValue(res.DefaultDB)
	} else {
		data.DefaultDB = types.StringNull()
	}
	if res.TagProvider != "" {
		data.TagProvider = types.StringValue(res.TagProvider)
	} else {
		data.TagProvider = types.StringNull()
	}
	if res.UserSource != "" {
		data.UserSource = types.StringValue(res.UserSource)
	} else {
		data.UserSource = types.StringNull()
	}
	if res.IdentityProvider != "" {
		data.IdentityProvider = types.StringValue(res.IdentityProvider)
	} else {
		data.IdentityProvider = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
