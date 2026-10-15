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
	_ resource.Resource                = &StoreAndForwardResource{}
	_ resource.ResourceWithImportState = &StoreAndForwardResource{}
)

func NewStoreAndForwardResource() resource.Resource {
	return &StoreAndForwardResource{}
}

// StoreAndForwardResource defines the resource implementation.
type StoreAndForwardResource struct {
	client  client.IgnitionClient
	generic base.GenericIgnitionResource[client.StoreAndForwardConfig, StoreAndForwardResourceModel]
}

// StoreAndForwardResourceModel describes the resource data model.
type StoreAndForwardResourceModel struct {
	base.BaseResourceModel
	TimeThresholdMs    types.Int64             `tfsdk:"time_threshold_ms"`
	ForwardRateMs      types.Int64             `tfsdk:"forward_rate_ms"`
	ForwardingPolicy   types.String            `tfsdk:"forwarding_policy"`
	ForwardingSchedule types.String            `tfsdk:"forwarding_schedule"`
	IsThirdParty       types.Bool              `tfsdk:"is_third_party"`
	DataThreshold      types.Int64             `tfsdk:"data_threshold"`
	BatchSize          types.Int64             `tfsdk:"batch_size"`
	ScanRateMs         types.Int64             `tfsdk:"scan_rate_ms"`
	PrimaryPolicy      *MaintenancePolicyModel `tfsdk:"primary_policy"`
	SecondaryPolicy    *MaintenancePolicyModel `tfsdk:"secondary_policy"`
}

type MaintenancePolicyModel struct {
	Action    types.String `tfsdk:"action"`
	LimitType types.String `tfsdk:"limit_type"`
	Value     types.Int64  `tfsdk:"value"`
}

func (r *StoreAndForwardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_store_forward"
}

func (r *StoreAndForwardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	maintenancePolicySchema := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Description: "What happens when the limit is reached (EVICT_OLDEST_DATA, PREVENT_NEW_DATA).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("EVICT_OLDEST_DATA", "PREVENT_NEW_DATA"),
				},
			},
			"limit_type": schema.StringAttribute{
				Description: "What the limit is measured in (TIME_DURATION, COUNT, FILE_SIZE).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("TIME_DURATION", "COUNT", "FILE_SIZE"),
				},
			},
			"value": schema.Int64Attribute{
				Description: "The limit, in the units of limit_type.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Manages a Store and Forward Engine in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the engine.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the engine.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the engine is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"time_threshold_ms": schema.Int64Attribute{
				Description: "Maximum time data accumulates in memory before writing to disk.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1000),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"forward_rate_ms": schema.Int64Attribute{
				Description: "Rate at which data is eligible for forwarding.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1000),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"forwarding_policy": schema.StringAttribute{
				Description: "Policy for forwarding (ALL, PRIMARY_ONLY, SECONDARY_ONLY).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ALL"),
				Validators: []validator.String{
					stringvalidator.OneOf("ALL", "PRIMARY_ONLY", "SECONDARY_ONLY"),
				},
			},
			"forwarding_schedule": schema.StringAttribute{
				Description: "Comma separated list of time ranges (e.g. 9:00-15:00).",
				Optional:    true,
			},
			"is_third_party": schema.BoolAttribute{
				Description: "Whether the engine is managed by a third party.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"data_threshold": schema.Int64Attribute{
				Description: "Maximum data amount in memory before writing to disk.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"batch_size": schema.Int64Attribute{
				Description: "Batch size for forwarding.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"scan_rate_ms": schema.Int64Attribute{
				Description: "Rate at which data pipelines are scanned.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1000),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"primary_policy": schema.SingleNestedAttribute{
				Description: "Maintenance policy for primary datastore (memory).",
				Optional:    true,
				Attributes:  maintenancePolicySchema.Attributes,
			},
			"secondary_policy": schema.SingleNestedAttribute{
				Description: "Maintenance policy for secondary datastore (disk).",
				Optional:    true,
				Attributes:  maintenancePolicySchema.Attributes,
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource.",
				Computed:    true,
			},
		},
	}
}

func (r *StoreAndForwardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.generic = base.GenericIgnitionResource[client.StoreAndForwardConfig, StoreAndForwardResourceModel]{
		Client:       c,
		Handler:      r,
		Module:       "ignition",
		ResourceType: "store-and-forward-engine",
		CreateFunc:   c.CreateStoreAndForward,
		GetFunc:      c.GetStoreAndForward,
		UpdateFunc:   c.UpdateStoreAndForward,
		DeleteFunc:   c.DeleteStoreAndForward,
	}
}

func (r *StoreAndForwardResource) MapPlanToClient(ctx context.Context, model *StoreAndForwardResourceModel) (client.StoreAndForwardConfig, error) {
	config := client.StoreAndForwardConfig{
		TimeThresholdMs:    int(model.TimeThresholdMs.ValueInt64()),
		ForwardRateMs:      int(model.ForwardRateMs.ValueInt64()),
		ForwardingPolicy:   model.ForwardingPolicy.ValueString(),
		ForwardingSchedule: model.ForwardingSchedule.ValueString(),
		IsThirdParty:       model.IsThirdParty.ValueBool(),
		DataThreshold:      int(model.DataThreshold.ValueInt64()),
		BatchSize:          int(model.BatchSize.ValueInt64()),
		ScanRateMs:         int(model.ScanRateMs.ValueInt64()),
	}

	if model.PrimaryPolicy != nil {
		config.PrimaryMaintenancePolicy = &client.StoreAndForwardMaintenancePolicy{
			Action: model.PrimaryPolicy.Action.ValueString(),
		}
		config.PrimaryMaintenancePolicy.Limit.LimitType = model.PrimaryPolicy.LimitType.ValueString()
		config.PrimaryMaintenancePolicy.Limit.Value = int(model.PrimaryPolicy.Value.ValueInt64())
	}

	if model.SecondaryPolicy != nil {
		config.SecondaryMaintenancePolicy = &client.StoreAndForwardMaintenancePolicy{
			Action: model.SecondaryPolicy.Action.ValueString(),
		}
		config.SecondaryMaintenancePolicy.Limit.LimitType = model.SecondaryPolicy.LimitType.ValueString()
		config.SecondaryMaintenancePolicy.Limit.Value = int(model.SecondaryPolicy.Value.ValueInt64())
	}

	return config, nil
}

func (r *StoreAndForwardResource) MapClientToState(ctx context.Context, name string, config *client.StoreAndForwardConfig, model *StoreAndForwardResourceModel) error {
	model.Name = types.StringValue(name)
	model.TimeThresholdMs = types.Int64Value(int64(config.TimeThresholdMs))
	model.ForwardRateMs = types.Int64Value(int64(config.ForwardRateMs))
	model.ForwardingPolicy = types.StringValue(config.ForwardingPolicy)
	model.ForwardingSchedule = base.GatewayString(model.ForwardingSchedule, config.ForwardingSchedule)
	model.IsThirdParty = types.BoolValue(config.IsThirdParty)
	model.DataThreshold = types.Int64Value(int64(config.DataThreshold))
	model.BatchSize = types.Int64Value(int64(config.BatchSize))
	model.ScanRateMs = types.Int64Value(int64(config.ScanRateMs))

	if config.PrimaryMaintenancePolicy != nil && model.PrimaryPolicy != nil {
		model.PrimaryPolicy = &MaintenancePolicyModel{
			Action:    types.StringValue(config.PrimaryMaintenancePolicy.Action),
			LimitType: types.StringValue(config.PrimaryMaintenancePolicy.Limit.LimitType),
			Value:     types.Int64Value(int64(config.PrimaryMaintenancePolicy.Limit.Value)),
		}
	} else {
		model.PrimaryPolicy = nil
	}

	if config.SecondaryMaintenancePolicy != nil && model.SecondaryPolicy != nil {
		model.SecondaryPolicy = &MaintenancePolicyModel{
			Action:    types.StringValue(config.SecondaryMaintenancePolicy.Action),
			LimitType: types.StringValue(config.SecondaryMaintenancePolicy.Limit.LimitType),
			Value:     types.Int64Value(int64(config.SecondaryMaintenancePolicy.Limit.Value)),
		}
	} else {
		model.SecondaryPolicy = nil
	}
	return nil
}

func (r *StoreAndForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StoreAndForwardResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *StoreAndForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StoreAndForwardResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *StoreAndForwardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StoreAndForwardResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *StoreAndForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StoreAndForwardResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *StoreAndForwardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &StoreAndForwardResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
