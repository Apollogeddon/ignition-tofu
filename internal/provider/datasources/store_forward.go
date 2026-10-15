package datasources

import (
	"context"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/apollogeddon/ignition-tofu/internal/provider/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &StoreAndForwardDataSource{}

func NewStoreAndForwardDataSource() datasource.DataSource {
	return &StoreAndForwardDataSource{}
}

// StoreAndForwardDataSource defines the data source implementation.
type StoreAndForwardDataSource struct {
	client client.IgnitionClient
}

// StoreAndForwardDataSourceModel describes the data source data model.
type StoreAndForwardDataSourceModel struct {
	Id                 types.String                      `tfsdk:"id"`
	Name               types.String                      `tfsdk:"name"`
	Description        types.String                      `tfsdk:"description"`
	Enabled            types.Bool                        `tfsdk:"enabled"`
	TimeThresholdMs    types.Int64                       `tfsdk:"time_threshold_ms"`
	ForwardRateMs      types.Int64                       `tfsdk:"forward_rate_ms"`
	ForwardingPolicy   types.String                      `tfsdk:"forwarding_policy"`
	ForwardingSchedule types.String                      `tfsdk:"forwarding_schedule"`
	IsThirdParty       types.Bool                        `tfsdk:"is_third_party"`
	DataThreshold      types.Int64                       `tfsdk:"data_threshold"`
	BatchSize          types.Int64                       `tfsdk:"batch_size"`
	ScanRateMs         types.Int64                       `tfsdk:"scan_rate_ms"`
	PrimaryPolicy      *resources.MaintenancePolicyModel `tfsdk:"primary_policy"`
	SecondaryPolicy    *resources.MaintenancePolicyModel `tfsdk:"secondary_policy"`
}

func (d *StoreAndForwardDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_store_forward"
}

func (d *StoreAndForwardDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	maintenancePolicySchema := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Description: "What happens when the limit is reached (EVICT_OLDEST_DATA, PREVENT_NEW_DATA).",
				Computed:    true,
			},
			"limit_type": schema.StringAttribute{
				Description: "What the limit is measured in (TIME_DURATION, COUNT, FILE_SIZE).",
				Computed:    true,
			},
			"value": schema.Int64Attribute{
				Description: "The limit, in the units of limit_type.",
				Computed:    true,
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Reads a Store and Forward Engine from Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the engine.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the engine.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the engine is enabled.",
				Computed:    true,
			},
			"time_threshold_ms": schema.Int64Attribute{
				Description: "Maximum time data accumulates in memory before writing to disk.",
				Computed:    true,
			},
			"forward_rate_ms": schema.Int64Attribute{
				Description: "Rate at which data is eligible for forwarding.",
				Computed:    true,
			},
			"forwarding_policy": schema.StringAttribute{
				Description: "Policy for forwarding (ALL, PRIMARY_ONLY, SECONDARY_ONLY).",
				Computed:    true,
			},
			"forwarding_schedule": schema.StringAttribute{
				Description: "Comma separated list of time ranges (e.g. 9:00-15:00).",
				Computed:    true,
			},
			"is_third_party": schema.BoolAttribute{
				Description: "Whether the engine is managed by a third party.",
				Computed:    true,
			},
			"data_threshold": schema.Int64Attribute{
				Description: "Maximum data amount in memory before writing to disk.",
				Computed:    true,
			},
			"batch_size": schema.Int64Attribute{
				Description: "Batch size for forwarding.",
				Computed:    true,
			},
			"scan_rate_ms": schema.Int64Attribute{
				Description: "Rate at which data pipelines are scanned.",
				Computed:    true,
			},
			"primary_policy": schema.SingleNestedAttribute{
				Description: "Maintenance policy for primary datastore (memory).",
				Computed:    true,
				Attributes:  maintenancePolicySchema.Attributes,
			},
			"secondary_policy": schema.SingleNestedAttribute{
				Description: "Maintenance policy for secondary datastore (disk).",
				Computed:    true,
				Attributes:  maintenancePolicySchema.Attributes,
			},
		},
	}
}

func (d *StoreAndForwardDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StoreAndForwardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		base.AddUnconfiguredError(&resp.Diagnostics)
		return
	}
	var data StoreAndForwardDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetStoreAndForward(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Store and Forward engine", err.Error())
		return
	}

	data.Id = types.StringValue(res.Name)
	if res.Enabled != nil {
		data.Enabled = types.BoolValue(*res.Enabled)
	} else {
		data.Enabled = types.BoolValue(true)
	}
	data.Description = types.StringValue(res.Description)

	data.TimeThresholdMs = types.Int64Value(int64(res.Config.TimeThresholdMs))
	data.ForwardRateMs = types.Int64Value(int64(res.Config.ForwardRateMs))
	data.ForwardingPolicy = types.StringValue(res.Config.ForwardingPolicy)
	data.ForwardingSchedule = types.StringValue(res.Config.ForwardingSchedule)
	data.IsThirdParty = types.BoolValue(res.Config.IsThirdParty)
	data.DataThreshold = types.Int64Value(int64(res.Config.DataThreshold))
	data.BatchSize = types.Int64Value(int64(res.Config.BatchSize))
	data.ScanRateMs = types.Int64Value(int64(res.Config.ScanRateMs))

	if res.Config.PrimaryMaintenancePolicy != nil {
		data.PrimaryPolicy = &resources.MaintenancePolicyModel{
			Action:    types.StringValue(res.Config.PrimaryMaintenancePolicy.Action),
			LimitType: types.StringValue(res.Config.PrimaryMaintenancePolicy.Limit.LimitType),
			Value:     types.Int64Value(int64(res.Config.PrimaryMaintenancePolicy.Limit.Value)),
		}
	}

	if res.Config.SecondaryMaintenancePolicy != nil {
		data.SecondaryPolicy = &resources.MaintenancePolicyModel{
			Action:    types.StringValue(res.Config.SecondaryMaintenancePolicy.Action),
			LimitType: types.StringValue(res.Config.SecondaryMaintenancePolicy.Limit.LimitType),
			Value:     types.Int64Value(int64(res.Config.SecondaryMaintenancePolicy.Limit.Value)),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
