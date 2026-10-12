package resources

import (
	"context"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &UserSourceResource{}
	_ resource.ResourceWithImportState = &UserSourceResource{}
)

func NewUserSourceResource() resource.Resource {
	return &UserSourceResource{}
}

// UserSourceResource defines the resource implementation.
type UserSourceResource struct {
	generic base.GenericIgnitionResource[client.UserSourceConfig, UserSourceResourceModel]
}

// UserSourceResourceModel describes the resource data model.
type UserSourceResourceModel struct {
	base.BaseResourceModel
	Type               types.String `tfsdk:"type"`
	FailoverProfile    types.String `tfsdk:"failover_profile"`
	FailoverMode       types.String `tfsdk:"failover_mode"`
	ScheduleRestricted types.Bool   `tfsdk:"schedule_restricted"`
}

func (r *UserSourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_source"
}

func (r *UserSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a User Source in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the user source.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of the user source (e.g., INTERNAL, ADEASY, ADHYBRID, AD_DB_HYBRID, DATASOURCE).",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the user source.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the user source is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"failover_profile": schema.StringAttribute{
				Description: "If this source is unreachable for authentication, this failover source will be used instead.",
				Optional:    true,
				Computed:    true,
			},
			"failover_mode": schema.StringAttribute{
				Description: "The failover mode to use if a failover source is set. Hard - failover only if this source is unreachable. Soft - try the failover source when a user fails to authenticate with this source.",
				Optional:    true,
				Computed:    true,
			},
			"schedule_restricted": schema.BoolAttribute{
				Description: "Users are only able to log in when their assigned schedule is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource, used for updates and deletes.",
				Computed:    true,
			},
		},
	}
}

func (r *UserSourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(client.IgnitionClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.IgnitionClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.generic = base.GenericIgnitionResource[client.UserSourceConfig, UserSourceResourceModel]{
		Client:       apiClient,
		Handler:      r,
		Module:       "ignition",
		ResourceType: "user-source",
		CreateFunc:   apiClient.CreateUserSource,
		GetFunc:      apiClient.GetUserSource,
		UpdateFunc:   apiClient.UpdateUserSource,
		DeleteFunc:   apiClient.DeleteUserSource,
	}
}

func (r *UserSourceResource) MapPlanToClient(ctx context.Context, model *UserSourceResourceModel) (client.UserSourceConfig, error) {
	profile := client.UserSourceProfile{
		Type: model.Type.ValueString(),
	}

	if !model.FailoverProfile.IsNull() {
		profile.FailoverProfile = model.FailoverProfile.ValueString()
	}
	if !model.FailoverMode.IsNull() {
		profile.FailoverMode = model.FailoverMode.ValueString()
	}
	if !model.ScheduleRestricted.IsNull() {
		profile.ScheduleRestricted = model.ScheduleRestricted.ValueBool()
	}

	return client.UserSourceConfig{
		Profile: profile,
	}, nil
}

func (r *UserSourceResource) MapClientToState(ctx context.Context, name string, config *client.UserSourceConfig, model *UserSourceResourceModel) error {
	model.Name = types.StringValue(name)
	model.Type = base.GatewayString(model.Type, config.Profile.Type)
	model.FailoverProfile = base.GatewayString(model.FailoverProfile, config.Profile.FailoverProfile)
	model.FailoverMode = base.GatewayString(model.FailoverMode, config.Profile.FailoverMode)

	model.ScheduleRestricted = types.BoolValue(config.Profile.ScheduleRestricted)
	return nil
}

func (r *UserSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserSourceResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *UserSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserSourceResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *UserSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserSourceResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *UserSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserSourceResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *UserSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &UserSourceResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
