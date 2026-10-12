package resources

import (
	"context"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &AuditProfileResource{}
	_ resource.ResourceWithImportState = &AuditProfileResource{}
)

func NewAuditProfileResource() resource.Resource {
	return &AuditProfileResource{}
}

// AuditProfileResource defines the resource implementation.
type AuditProfileResource struct {
	generic base.GenericIgnitionResource[client.AuditProfileConfig, AuditProfileResourceModel]
}

// AuditProfileResourceModel describes the resource data model.
type AuditProfileResourceModel struct {
	base.BaseResourceModel
	Type                  types.String `tfsdk:"type"`
	RetentionDays         types.Int64  `tfsdk:"retention_days"`
	Database              types.String `tfsdk:"database"`
	PruneEnabled          types.Bool   `tfsdk:"prune_enabled"`
	AutoCreate            types.Bool   `tfsdk:"auto_create"`
	TableName             types.String `tfsdk:"table_name"`
	RemoteServer          types.String `tfsdk:"remote_server"`
	RemoteProfile         types.String `tfsdk:"remote_profile"`
	EnableStoreAndForward types.Bool   `tfsdk:"enable_store_and_forward"`
}

func (r *AuditProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_profile"
}

func (r *AuditProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Audit Profile in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the audit profile.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the audit profile.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the audit profile is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The type of the audit profile (database, remote, edge, local).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("database", "remote", "edge", "local"),
				},
			},
			"retention_days": schema.Int64Attribute{
				Description: "How long (in days) should audit records be kept? If set to 0, pruning will be disabled. Default: 90.",
				Optional:    true,
				Computed:    true,
			},
			"database": schema.StringAttribute{
				Description: "The database connection to use to store audit events (for 'database' type).",
				Optional:    true,
			},
			"prune_enabled": schema.BoolAttribute{
				Description: "If false, this audit profile will never prune records (for 'database' type). Default: false.",
				Optional:    true,
				Computed:    true,
			},
			"auto_create": schema.BoolAttribute{
				Description: "If true, the table schema will be automatically created (for 'database' type). Default: true.",
				Optional:    true,
				Computed:    true,
			},
			"table_name": schema.StringAttribute{
				Description: "The name of the table to store audit events (for 'database' type). Default: 'audit_events'.",
				Optional:    true,
				Computed:    true,
			},
			"remote_server": schema.StringAttribute{
				Description: "The remote system to send audit events to (for 'remote' type).",
				Optional:    true,
			},
			"remote_profile": schema.StringAttribute{
				Description: "The audit profile on the remote system to log events into (for 'remote' type).",
				Optional:    true,
			},
			"enable_store_and_forward": schema.BoolAttribute{
				Description: "If enabled, audit events will be stored through the store and forward system (for 'remote' type).",
				Optional:    true,
				Computed:    true,
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource, used for updates and deletes.",
				Computed:    true,
			},
		},
	}
}

func (r *AuditProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.generic = base.GenericIgnitionResource[client.AuditProfileConfig, AuditProfileResourceModel]{
		Client:       apiClient,
		Handler:      r,
		Module:       "ignition",
		ResourceType: "audit-profile",
		CreateFunc:   apiClient.CreateAuditProfile,
		GetFunc:      apiClient.GetAuditProfile,
		UpdateFunc:   apiClient.UpdateAuditProfile,
		DeleteFunc:   apiClient.DeleteAuditProfile,
	}
}

func (r *AuditProfileResource) MapPlanToClient(ctx context.Context, model *AuditProfileResourceModel) (client.AuditProfileConfig, error) {
	profile := client.AuditProfileProfile{
		Type: model.Type.ValueString(),
	}
	if !model.RetentionDays.IsNull() {
		profile.RetentionDays = int(model.RetentionDays.ValueInt64())
	}

	settings := client.AuditProfileSettings{}
	if !model.Database.IsNull() {
		settings.DatabaseName = model.Database.ValueString()
	}
	if !model.PruneEnabled.IsNull() {
		settings.PruneEnabled = model.PruneEnabled.ValueBool()
	}
	if !model.AutoCreate.IsNull() {
		settings.AutoCreate = model.AutoCreate.ValueBool()
	}
	if !model.TableName.IsNull() {
		settings.TableName = model.TableName.ValueString()
	}
	if !model.RemoteServer.IsNull() {
		settings.RemoteServer = model.RemoteServer.ValueString()
	}
	if !model.RemoteProfile.IsNull() {
		settings.RemoteProfile = model.RemoteProfile.ValueString()
	}
	if !model.EnableStoreAndForward.IsNull() {
		settings.EnableStoreAndForward = model.EnableStoreAndForward.ValueBool()
	}

	return client.AuditProfileConfig{
		Profile:  profile,
		Settings: settings,
	}, nil
}

func (r *AuditProfileResource) MapClientToState(ctx context.Context, name string, config *client.AuditProfileConfig, model *AuditProfileResourceModel) error {
	model.Name = types.StringValue(name)

	model.Type = types.StringValue(config.Profile.Type)
	model.RetentionDays = types.Int64Value(int64(config.Profile.RetentionDays))

	model.Database = base.GatewayString(model.Database, config.Settings.DatabaseName)

	model.PruneEnabled = types.BoolValue(config.Settings.PruneEnabled)
	model.AutoCreate = types.BoolValue(config.Settings.AutoCreate)

	// A database profile without a table name writes to the default table, so
	// that is what a name cleared on the gateway shows as.
	if config.Settings.TableName == "" && model.Type.ValueString() == "database" {
		model.TableName = types.StringValue("audit_events")
	} else {
		model.TableName = base.GatewayString(model.TableName, config.Settings.TableName)
	}

	model.RemoteServer = base.GatewayString(model.RemoteServer, config.Settings.RemoteServer)
	model.RemoteProfile = base.GatewayString(model.RemoteProfile, config.Settings.RemoteProfile)

	model.EnableStoreAndForward = types.BoolValue(config.Settings.EnableStoreAndForward)

	return nil
}

func (r *AuditProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AuditProfileResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AuditProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AuditProfileResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AuditProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AuditProfileResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AuditProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AuditProfileResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AuditProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &AuditProfileResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
