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
	_ resource.Resource                = &AlarmJournalResource{}
	_ resource.ResourceWithImportState = &AlarmJournalResource{}
)

func NewAlarmJournalResource() resource.Resource {
	return &AlarmJournalResource{}
}

// AlarmJournalResource defines the resource implementation.
type AlarmJournalResource struct {
	client  client.IgnitionClient
	generic base.GenericIgnitionResource[client.AlarmJournalConfig, AlarmJournalResourceModel]
}

// AlarmJournalResourceModel describes the resource data model.
type AlarmJournalResourceModel struct {
	base.BaseResourceModel
	Type          types.String `tfsdk:"type"`
	Datasource    types.String `tfsdk:"datasource"`
	TableName     types.String `tfsdk:"table_name"`
	MinPriority   types.String `tfsdk:"min_priority"`
	TargetServer  types.String `tfsdk:"target_server"`
	TargetJournal types.String `tfsdk:"target_journal"`
}

func (r *AlarmJournalResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alarm_journal"
}

func (r *AlarmJournalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Alarm Journal in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the alarm journal.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the alarm journal.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alarm journal is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The type of the alarm journal (DATASOURCE, REMOTE, LOCAL).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("DATASOURCE", "REMOTE", "LOCAL"),
				},
			},
			"datasource": schema.StringAttribute{
				Description: "The database connection to use (for DATASOURCE type).",
				Optional:    true,
			},
			"table_name": schema.StringAttribute{
				Description: "The table name to store events (for DATASOURCE type).",
				Optional:    true,
				Computed:    true,
			},
			"min_priority": schema.StringAttribute{
				Description: "The minimum priority of events to store (Diagnostic, Low, Medium, High, Critical).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Diagnostic", "Low", "Medium", "High", "Critical"),
				},
			},
			"target_server": schema.StringAttribute{
				Description: "The remote gateway name (for REMOTE type).",
				Optional:    true,
			},
			"target_journal": schema.StringAttribute{
				Description: "The alarm journal on the remote gateway (for REMOTE type).",
				Optional:    true,
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource, used for updates and deletes.",
				Computed:    true,
			},
		},
	}
}

func (r *AlarmJournalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(client.IgnitionClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.IgnitionClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
	r.generic = base.GenericIgnitionResource[client.AlarmJournalConfig, AlarmJournalResourceModel]{
		Client:       c,
		Handler:      r,
		Module:       "ignition",
		ResourceType: "alarm-journal",
		CreateFunc:   c.CreateAlarmJournal,
		GetFunc:      c.GetAlarmJournal,
		UpdateFunc:   c.UpdateAlarmJournal,
		DeleteFunc:   c.DeleteAlarmJournal,
	}
}

func (r *AlarmJournalResource) MapPlanToClient(ctx context.Context, model *AlarmJournalResourceModel) (client.AlarmJournalConfig, error) {
	settings := client.AlarmJournalSettings{}

	if model.Type.ValueString() == "DATASOURCE" {
		if !model.Datasource.IsNull() {
			settings.Datasource = model.Datasource.ValueString()
		}

		settings.Advanced = &struct {
			TableName          string `json:"tableName,omitempty"`
			DataTableName      string `json:"dataTableName,omitempty"`
			UseStoreAndForward bool   `json:"useStoreAndForward,omitempty"`
		}{}

		if !model.TableName.IsNull() {
			settings.Advanced.TableName = model.TableName.ValueString()
		}

		settings.Events = &struct {
			MinPriority            string `json:"minPriority,omitempty"`
			StoreShelvedEvents     bool   `json:"storeShelvedEvents,omitempty"`
			StoreFromEnabledChange bool   `json:"storeFromEnabledChange,omitempty"`
		}{}

		if !model.MinPriority.IsNull() {
			settings.Events.MinPriority = model.MinPriority.ValueString()
		}
	} else if model.Type.ValueString() == "REMOTE" {
		settings.RemoteGateway = &struct {
			TargetServer  string `json:"targetServer,omitempty"`
			TargetJournal string `json:"targetJournal,omitempty"`
		}{}
		if !model.TargetServer.IsNull() {
			settings.RemoteGateway.TargetServer = model.TargetServer.ValueString()
		}
		if !model.TargetJournal.IsNull() {
			settings.RemoteGateway.TargetJournal = model.TargetJournal.ValueString()
		}
	}

	return client.AlarmJournalConfig{
		Profile: client.AlarmJournalProfile{
			Type: model.Type.ValueString(),
		},
		Settings: settings,
	}, nil
}

func (r *AlarmJournalResource) MapClientToState(ctx context.Context, name string, config *client.AlarmJournalConfig, model *AlarmJournalResourceModel) error {
	model.Name = types.StringValue(name)

	model.Type = base.GatewayString(model.Type, config.Profile.Type)

	model.Datasource = types.StringNull()
	model.TableName = types.StringNull()
	model.MinPriority = types.StringNull()
	model.TargetServer = types.StringNull()
	model.TargetJournal = types.StringNull()

	switch config.Profile.Type {
	case "DATASOURCE":
		model.Datasource = types.StringValue(config.Settings.Datasource)
		if config.Settings.Advanced != nil {
			model.TableName = types.StringValue(config.Settings.Advanced.TableName)
		}
		if config.Settings.Events != nil {
			model.MinPriority = types.StringValue(config.Settings.Events.MinPriority)
		}
	case "REMOTE":
		if config.Settings.RemoteGateway != nil {
			model.TargetServer = types.StringValue(config.Settings.RemoteGateway.TargetServer)
			model.TargetJournal = types.StringValue(config.Settings.RemoteGateway.TargetJournal)
		}
	}
	return nil
}

func (r *AlarmJournalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AlarmJournalResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AlarmJournalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AlarmJournalResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AlarmJournalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AlarmJournalResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AlarmJournalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AlarmJournalResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AlarmJournalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &AlarmJournalResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
