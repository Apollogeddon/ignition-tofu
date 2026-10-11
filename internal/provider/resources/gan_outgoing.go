package resources

import (
	"context"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &GanOutgoingResource{}
	_ resource.ResourceWithImportState = &GanOutgoingResource{}
)

func NewGanOutgoingResource() resource.Resource {
	return &GanOutgoingResource{}
}

// GanOutgoingResource defines the resource implementation.
type GanOutgoingResource struct {
	client  client.IgnitionClient
	generic base.GenericIgnitionResource[client.GanOutgoingConfig, GanOutgoingResourceModel]
}

// GanOutgoingResourceModel describes the resource data model.
type GanOutgoingResourceModel struct {
	base.BaseResourceModel
	Host                     types.String  `tfsdk:"host"`
	Port                     types.Int64   `tfsdk:"port"`
	UseSSL                   types.Bool    `tfsdk:"use_ssl"`
	PingRateMillis           types.Float64 `tfsdk:"ping_rate_millis"`
	PingTimeoutMillis        types.Float64 `tfsdk:"ping_timeout_millis"`
	PingMaxMissed            types.Float64 `tfsdk:"ping_max_missed"`
	WsTimeoutMillis          types.Float64 `tfsdk:"ws_timeout_millis"`
	HttpConnectTimeoutMillis types.Float64 `tfsdk:"http_connect_timeout_millis"`
	HttpReadTimeoutMillis    types.Float64 `tfsdk:"http_read_timeout_millis"`
	SendThreads              types.Float64 `tfsdk:"send_threads"`
	ReceiveThreads           types.Float64 `tfsdk:"receive_threads"`
}

func (r *GanOutgoingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gan_outgoing"
}

func (r *GanOutgoingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Outgoing Gateway Network Connection in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the connection.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the connection.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the connection is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"host": schema.StringAttribute{
				Description: "The address of the remote server.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port of the remote server.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(8060),
			},
			"use_ssl": schema.BoolAttribute{
				Description: "Whether to use SSL for the connection.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ping_rate_millis": schema.Float64Attribute{
				Description: "How often, in milliseconds, the connection pings the remote gateway.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(2000),
			},
			"ping_timeout_millis": schema.Float64Attribute{
				Description: "How long, in milliseconds, a ping waits for its reply.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(60000),
			},
			"ping_max_missed": schema.Float64Attribute{
				Description: "How many pings in a row may go unanswered before the connection is considered down.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(3),
			},
			"ws_timeout_millis": schema.Float64Attribute{
				Description: "The websocket timeout, in milliseconds.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(10000),
			},
			"http_connect_timeout_millis": schema.Float64Attribute{
				Description: "How long, in milliseconds, to wait for an HTTP connection to the remote gateway to open.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(10000),
			},
			"http_read_timeout_millis": schema.Float64Attribute{
				Description: "How long, in milliseconds, to wait for an HTTP response from the remote gateway.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(30000),
			},
			"send_threads": schema.Float64Attribute{
				Description: "How many threads send messages over the connection.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(1),
			},
			"receive_threads": schema.Float64Attribute{
				Description: "How many threads receive messages from the connection.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(1),
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource.",
				Computed:    true,
			},
		},
	}
}

func (r *GanOutgoingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.generic = base.GenericIgnitionResource[client.GanOutgoingConfig, GanOutgoingResourceModel]{
		Client:       c,
		Handler:      r,
		Module:       "ignition",
		ResourceType: "gateway-network-outgoing",
		CreateFunc:   c.CreateGanOutgoing,
		GetFunc:      c.GetGanOutgoing,
		UpdateFunc:   c.UpdateGanOutgoing,
		DeleteFunc:   c.DeleteGanOutgoing,
	}
}

func (r *GanOutgoingResource) MapPlanToClient(ctx context.Context, model *GanOutgoingResourceModel) (client.GanOutgoingConfig, error) {
	return client.GanOutgoingConfig{
		Host:                     model.Host.ValueString(),
		Port:                     int(model.Port.ValueInt64()),
		UseSSL:                   model.UseSSL.ValueBool(),
		PingRateMillis:           model.PingRateMillis.ValueFloat64(),
		PingTimeoutMillis:        model.PingTimeoutMillis.ValueFloat64(),
		PingMaxMissed:            model.PingMaxMissed.ValueFloat64(),
		WsTimeoutMillis:          model.WsTimeoutMillis.ValueFloat64(),
		HttpConnectTimeoutMillis: model.HttpConnectTimeoutMillis.ValueFloat64(),
		HttpReadTimeoutMillis:    model.HttpReadTimeoutMillis.ValueFloat64(),
		SendThreads:              model.SendThreads.ValueFloat64(),
		ReceiveThreads:           model.ReceiveThreads.ValueFloat64(),
	}, nil
}

func (r *GanOutgoingResource) MapClientToState(ctx context.Context, name string, config *client.GanOutgoingConfig, model *GanOutgoingResourceModel) error {
	model.Name = types.StringValue(name)
	model.Host = types.StringValue(config.Host)
	model.Port = types.Int64Value(int64(config.Port))
	model.UseSSL = types.BoolValue(config.UseSSL)
	model.PingRateMillis = types.Float64Value(config.PingRateMillis)
	model.PingTimeoutMillis = types.Float64Value(config.PingTimeoutMillis)
	model.PingMaxMissed = types.Float64Value(config.PingMaxMissed)
	model.WsTimeoutMillis = types.Float64Value(config.WsTimeoutMillis)
	model.HttpConnectTimeoutMillis = types.Float64Value(config.HttpConnectTimeoutMillis)
	model.HttpReadTimeoutMillis = types.Float64Value(config.HttpReadTimeoutMillis)
	model.SendThreads = types.Float64Value(config.SendThreads)
	model.ReceiveThreads = types.Float64Value(config.ReceiveThreads)
	return nil
}

func (r *GanOutgoingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GanOutgoingResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *GanOutgoingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GanOutgoingResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *GanOutgoingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GanOutgoingResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *GanOutgoingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GanOutgoingResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *GanOutgoingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &GanOutgoingResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
