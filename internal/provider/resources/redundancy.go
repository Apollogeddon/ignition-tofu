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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RedundancyResource{}

func NewRedundancyResource() resource.Resource {
	return &RedundancyResource{}
}

// RedundancyResource defines the resource implementation.
type RedundancyResource struct {
	client  client.IgnitionClient
	generic base.GenericIgnitionResource[client.RedundancyConfig, RedundancyResourceModel]
}

// RedundancyResourceModel describes the resource data model.
type RedundancyResourceModel struct {
	base.BaseResourceModel
	Role                types.String         `tfsdk:"role"`
	ActiveHistoryLevel  types.String         `tfsdk:"active_history_level"`
	JoinWaitTime        types.Int64          `tfsdk:"join_wait_time"`
	RecoveryMode        types.String         `tfsdk:"recovery_mode"`
	AllowHistoryCleanup types.Bool           `tfsdk:"allow_history_cleanup"`
	GatewayNetworkSetup *GatewayNetworkSetup `tfsdk:"gateway_network_setup"`
}

type GatewayNetworkSetup struct {
	Host               types.String  `tfsdk:"host"`
	Port               types.Int64   `tfsdk:"port"`
	EnableSsl          types.Bool    `tfsdk:"enable_ssl"`
	PingRate           types.Float64 `tfsdk:"ping_rate"`
	PingTimeout        types.Float64 `tfsdk:"ping_timeout"`
	PingMaxMissed      types.Float64 `tfsdk:"ping_max_missed"`
	WebsocketTimeout   types.Float64 `tfsdk:"websocket_timeout"`
	HttpConnectTimeout types.Float64 `tfsdk:"http_connect_timeout"`
	HttpReadTimeout    types.Float64 `tfsdk:"http_read_timeout"`
	SendThreads        types.Float64 `tfsdk:"send_threads"`
	ReceiveThreads     types.Float64 `tfsdk:"receive_threads"`
}

func (r *RedundancyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redundancy"
}

func (r *RedundancyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Gateway Redundancy Settings. This is a singleton resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Internal name for the resource (fixed to 'gateway-redundancy').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("gateway-redundancy"),
			},
			"description": schema.StringAttribute{
				Description: "The description of the settings.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether redundancy is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"role": schema.StringAttribute{
				Description: "The node's redundancy role (Independent, Backup, Master).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Independent", "Backup", "Master"),
				},
			},
			"active_history_level": schema.StringAttribute{
				Description: "The redundancy setting's active history level (Partial, Full).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Full"),
				Validators: []validator.String{
					stringvalidator.OneOf("Partial", "Full"),
				},
			},
			"join_wait_time": schema.Int64Attribute{
				Description: "How long, in milliseconds, a node waits on startup to reach its peer before it decides its state.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10000),
			},
			"recovery_mode": schema.StringAttribute{
				Description: "How control returns to the master after a failover, e.g. Automatic or Manual.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Automatic"),
			},
			"allow_history_cleanup": schema.BoolAttribute{
				Description: "The redundancy setting that allows history cleanup.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway_network_setup": schema.SingleNestedAttribute{
				Description: "Gateway network settings to establish a connection to the redundant master. (Only applies to Backup)",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						Description: "The address of the redundant master.",
						Optional:    true,
					},
					"port": schema.Int64Attribute{
						Description: "The gateway network port of the redundant master.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(8060),
					},
					"enable_ssl": schema.BoolAttribute{
						Description: "Whether to use SSL for the connection to the master.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"ping_rate":            schema.Float64Attribute{Description: "How often, in milliseconds, the connection pings the master.", Optional: true, Computed: true, Default: float64default.StaticFloat64(2000)},
					"ping_timeout":         schema.Float64Attribute{Description: "How long, in milliseconds, a ping waits for its reply.", Optional: true, Computed: true, Default: float64default.StaticFloat64(60000)},
					"ping_max_missed":      schema.Float64Attribute{Description: "How many pings in a row may go unanswered before the connection is considered down.", Optional: true, Computed: true, Default: float64default.StaticFloat64(3)},
					"websocket_timeout":    schema.Float64Attribute{Description: "The websocket timeout, in milliseconds.", Optional: true, Computed: true, Default: float64default.StaticFloat64(10000)},
					"http_connect_timeout": schema.Float64Attribute{Description: "How long, in milliseconds, to wait for an HTTP connection to the master to open.", Optional: true, Computed: true, Default: float64default.StaticFloat64(10000)},
					"http_read_timeout":    schema.Float64Attribute{Description: "How long, in milliseconds, to wait for an HTTP response from the master.", Optional: true, Computed: true, Default: float64default.StaticFloat64(30000)},
					"send_threads":         schema.Float64Attribute{Description: "How many threads send messages over the connection.", Optional: true, Computed: true, Default: float64default.StaticFloat64(1)},
					"receive_threads":      schema.Float64Attribute{Description: "How many threads receive messages from the connection.", Optional: true, Computed: true, Default: float64default.StaticFloat64(1)},
				},
			},
		},
	}
}

func (r *RedundancyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.generic = base.GenericIgnitionResource[client.RedundancyConfig, RedundancyResourceModel]{
		Client:       c,
		Handler:      r,
		ResourceType: "gateway-redundancy",
		CreateFunc: func(ctx context.Context, res client.ResourceResponse[client.RedundancyConfig]) (*client.ResourceResponse[client.RedundancyConfig], error) {
			err := c.UpdateRedundancyConfig(ctx, res.Config)
			if err != nil {
				return nil, err
			}
			return &client.ResourceResponse[client.RedundancyConfig]{
				Name:   "gateway-redundancy",
				Config: res.Config,
			}, nil
		},
		GetFunc: func(ctx context.Context, _ string) (*client.ResourceResponse[client.RedundancyConfig], error) {
			conf, err := c.GetRedundancyConfig(ctx)
			if err != nil {
				return nil, err
			}
			return &client.ResourceResponse[client.RedundancyConfig]{
				Name:   "gateway-redundancy",
				Config: *conf,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, res client.ResourceResponse[client.RedundancyConfig]) (*client.ResourceResponse[client.RedundancyConfig], error) {
			err := c.UpdateRedundancyConfig(ctx, res.Config)
			if err != nil {
				return nil, err
			}
			return &client.ResourceResponse[client.RedundancyConfig]{
				Name:   "gateway-redundancy",
				Config: res.Config,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, _, _ string) error {
			// The gateway rejects a zero-valued JoinWaitTime, so reset to the
			// same defaults the schema uses rather than an all-zero config.
			return c.UpdateRedundancyConfig(ctx, client.RedundancyConfig{
				Role:               "Independent",
				ActiveHistoryLevel: "Full",
				JoinWaitTime:       10000,
				RecoveryMode:       "Automatic",
			})
		},
	}
}

func (r *RedundancyResource) MapPlanToClient(ctx context.Context, model *RedundancyResourceModel) (client.RedundancyConfig, error) {
	config := client.RedundancyConfig{
		Role:                model.Role.ValueString(),
		ActiveHistoryLevel:  model.ActiveHistoryLevel.ValueString(),
		JoinWaitTime:        int(model.JoinWaitTime.ValueInt64()),
		RecoveryMode:        model.RecoveryMode.ValueString(),
		AllowHistoryCleanup: model.AllowHistoryCleanup.ValueBool(),
	}

	if model.GatewayNetworkSetup != nil {
		config.GatewayNetworkSetup = &struct {
			Host               string  `json:"host,omitempty"`
			Port               int     `json:"port,omitempty"`
			EnableSsl          bool    `json:"enableSsl,omitempty"`
			PingRate           float64 `json:"pingRate,omitempty"`
			PingTimeout        float64 `json:"pingTimeout,omitempty"`
			PingMaxMissed      float64 `json:"pingMaxMissed,omitempty"`
			WebsocketTimeout   float64 `json:"websocketTimeout,omitempty"`
			HttpConnectTimeout float64 `json:"httpConnectTimeout,omitempty"`
			HttpReadTimeout    float64 `json:"httpReadTimeout,omitempty"`
			SendThreads        float64 `json:"sendThreads,omitempty"`
			ReceiveThreads     float64 `json:"receiveThreads,omitempty"`
		}{
			Host:               model.GatewayNetworkSetup.Host.ValueString(),
			Port:               int(model.GatewayNetworkSetup.Port.ValueInt64()),
			EnableSsl:          model.GatewayNetworkSetup.EnableSsl.ValueBool(),
			PingRate:           model.GatewayNetworkSetup.PingRate.ValueFloat64(),
			PingTimeout:        model.GatewayNetworkSetup.PingTimeout.ValueFloat64(),
			PingMaxMissed:      model.GatewayNetworkSetup.PingMaxMissed.ValueFloat64(),
			WebsocketTimeout:   model.GatewayNetworkSetup.WebsocketTimeout.ValueFloat64(),
			HttpConnectTimeout: model.GatewayNetworkSetup.HttpConnectTimeout.ValueFloat64(),
			HttpReadTimeout:    model.GatewayNetworkSetup.HttpReadTimeout.ValueFloat64(),
			SendThreads:        model.GatewayNetworkSetup.SendThreads.ValueFloat64(),
			ReceiveThreads:     model.GatewayNetworkSetup.ReceiveThreads.ValueFloat64(),
		}
	}

	return config, nil
}

func (r *RedundancyResource) MapClientToState(ctx context.Context, name string, config *client.RedundancyConfig, model *RedundancyResourceModel) error {
	model.Name = types.StringValue(name)
	model.Role = types.StringValue(config.Role)
	model.ActiveHistoryLevel = types.StringValue(config.ActiveHistoryLevel)
	model.JoinWaitTime = types.Int64Value(int64(config.JoinWaitTime))
	model.RecoveryMode = types.StringValue(config.RecoveryMode)
	model.AllowHistoryCleanup = types.BoolValue(config.AllowHistoryCleanup)

	// gateway_network_setup is Optional (not Computed): the gateway always
	// returns some populated struct here regardless of role, but Terraform
	// only allows the provider to report a value for this attribute if the
	// user actually configured it.
	if config.GatewayNetworkSetup != nil && model.GatewayNetworkSetup != nil {
		model.GatewayNetworkSetup = &GatewayNetworkSetup{
			Host:               types.StringValue(config.GatewayNetworkSetup.Host),
			Port:               types.Int64Value(int64(config.GatewayNetworkSetup.Port)),
			EnableSsl:          types.BoolValue(config.GatewayNetworkSetup.EnableSsl),
			PingRate:           types.Float64Value(config.GatewayNetworkSetup.PingRate),
			PingTimeout:        types.Float64Value(config.GatewayNetworkSetup.PingTimeout),
			PingMaxMissed:      types.Float64Value(config.GatewayNetworkSetup.PingMaxMissed),
			WebsocketTimeout:   types.Float64Value(config.GatewayNetworkSetup.WebsocketTimeout),
			HttpConnectTimeout: types.Float64Value(config.GatewayNetworkSetup.HttpConnectTimeout),
			HttpReadTimeout:    types.Float64Value(config.GatewayNetworkSetup.HttpReadTimeout),
			SendThreads:        types.Float64Value(config.GatewayNetworkSetup.SendThreads),
			ReceiveThreads:     types.Float64Value(config.GatewayNetworkSetup.ReceiveThreads),
		}
	}
	return nil
}

func (r *RedundancyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RedundancyResourceModel
	// Ensure fixed name
	data.Name = types.StringValue("gateway-redundancy")
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *RedundancyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RedundancyResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *RedundancyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RedundancyResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *RedundancyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RedundancyResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}
