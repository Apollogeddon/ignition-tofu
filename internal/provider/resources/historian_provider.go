package resources

import (
	"context"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &HistorianProviderResource{}
	_ resource.ResourceWithImportState = &HistorianProviderResource{}
)

// The gateway's defaults for the settings the resource doesn't expose.
const (
	historianSeedQueryLimit     = 2
	historianOptimizedWindowSec = 60
	historianStaleMultiplier    = 2
)

var historianUnits = []string{"HOUR", "DAY", "WEEK", "MONTH", "YEAR"}

func NewHistorianProviderResource() resource.Resource {
	return &HistorianProviderResource{}
}

// HistorianProviderResource manages an Ignition 8.3 historian provider.
type HistorianProviderResource struct {
	client  client.IgnitionClient
	generic base.GenericIgnitionResource[client.HistorianProviderConfig, HistorianProviderResourceModel]
}

// HistorianProviderResourceModel describes the resource data model.
type HistorianProviderResourceModel struct {
	base.BaseResourceModel
	Type             types.String `tfsdk:"type"`
	Database         types.String `tfsdk:"database"`
	PartitionEnabled types.Bool   `tfsdk:"partition_enabled"`
	PartitionSize    types.Int64  `tfsdk:"partition_size"`
	PartitionUnits   types.String `tfsdk:"partition_units"`
	PruningEnabled   types.Bool   `tfsdk:"pruning_enabled"`
	PruningAge       types.Int64  `tfsdk:"pruning_age"`
	PruningUnits     types.String `tfsdk:"pruning_units"`
}

func (r *HistorianProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_historian_provider"
}

func (r *HistorianProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a historian provider in Ignition 8.3, where tag history is stored. " +
			"Supports the SQL Historian, which stores history in a database connection; " +
			"8.1 creates one for each database connection itself.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the historian provider: tags name it in historyProvider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the historian provider.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the historian provider is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The historian's type: SqlHistorian.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("SqlHistorian"),
				Validators: []validator.String{
					stringvalidator.OneOf("SqlHistorian"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database": schema.StringAttribute{
				Description: "The database connection the history is stored in.",
				Required:    true,
			},
			"partition_enabled": schema.BoolAttribute{
				Description: "Whether the history is split into a table per partition period.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"partition_size": schema.Int64Attribute{
				Description: "The length of a partition, in partition_units.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"partition_units": schema.StringAttribute{
				Description: "The unit of partition_size: HOUR, DAY, WEEK, MONTH or YEAR.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("MONTH"),
				Validators:  []validator.String{stringvalidator.OneOf(historianUnits...)},
			},
			"pruning_enabled": schema.BoolAttribute{
				Description: "Whether partitions older than pruning_age are deleted.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"pruning_age": schema.Int64Attribute{
				Description: "The age, in pruning_units, past which a partition is deleted.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"pruning_units": schema.StringAttribute{
				Description: "The unit of pruning_age: HOUR, DAY, WEEK, MONTH or YEAR.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("YEAR"),
				Validators:  []validator.String{stringvalidator.OneOf(historianUnits...)},
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource, used for updates and deletes.",
				Computed:    true,
			},
		},
	}
}

func (r *HistorianProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.generic = base.GenericIgnitionResource[client.HistorianProviderConfig, HistorianProviderResourceModel]{
		Client:       c,
		Handler:      r,
		Module:       "com.inductiveautomation.historian",
		ResourceType: "historian-provider",
		CreateFunc:   c.CreateHistorianProvider,
		GetFunc:      c.GetHistorianProvider,
		UpdateFunc:   c.UpdateHistorianProvider,
		DeleteFunc:   c.DeleteHistorianProvider,
	}
}

func (r *HistorianProviderResource) MapPlanToClient(ctx context.Context, model *HistorianProviderResourceModel) (client.HistorianProviderConfig, error) {
	return client.HistorianProviderConfig{
		Profile: client.HistorianProviderProfile{Type: model.Type.ValueString()},
		Settings: client.HistorianProviderSettings{
			Database: model.Database.ValueString(),
			Partition: client.HistorianPartition{
				Enabled:                 model.PartitionEnabled.ValueBool(),
				Size:                    model.PartitionSize.ValueInt64(),
				SizeUnits:               model.PartitionUnits.ValueString(),
				PartitionSeedQueryLimit: historianSeedQueryLimit,
				OptimizedWindowSeconds:  historianOptimizedWindowSec,
			},
			Pruning: client.HistorianPruning{
				Enabled:  model.PruningEnabled.ValueBool(),
				Age:      model.PruningAge.ValueInt64(),
				AgeUnits: model.PruningUnits.ValueString(),
			},
			TrackSce:        true,
			StaleMultiplier: historianStaleMultiplier,
		},
	}, nil
}

func (r *HistorianProviderResource) MapClientToState(ctx context.Context, name string, config *client.HistorianProviderConfig, model *HistorianProviderResourceModel) error {
	model.Name = types.StringValue(name)
	model.Type = base.GatewayString(model.Type, config.Profile.Type)
	model.Database = types.StringValue(config.Settings.Database)
	model.PartitionEnabled = types.BoolValue(config.Settings.Partition.Enabled)
	model.PartitionSize = types.Int64Value(config.Settings.Partition.Size)
	model.PartitionUnits = types.StringValue(config.Settings.Partition.SizeUnits)
	model.PruningEnabled = types.BoolValue(config.Settings.Pruning.Enabled)
	model.PruningAge = types.Int64Value(config.Settings.Pruning.Age)
	model.PruningUnits = types.StringValue(config.Settings.Pruning.AgeUnits)
	return nil
}

func (r *HistorianProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HistorianProviderResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *HistorianProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HistorianProviderResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *HistorianProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HistorianProviderResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *HistorianProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HistorianProviderResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *HistorianProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &HistorianProviderResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
