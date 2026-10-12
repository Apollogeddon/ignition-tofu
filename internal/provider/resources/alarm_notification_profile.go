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
	_ resource.Resource                = &AlarmNotificationProfileResource{}
	_ resource.ResourceWithImportState = &AlarmNotificationProfileResource{}
)

func NewAlarmNotificationProfileResource() resource.Resource {
	return &AlarmNotificationProfileResource{}
}

// AlarmNotificationProfileResource defines the resource implementation.
type AlarmNotificationProfileResource struct {
	generic base.GenericIgnitionResource[client.AlarmNotificationProfileConfig, AlarmNotificationProfileResourceModel]
}

// AlarmNotificationProfileResourceModel describes the resource data model.
type AlarmNotificationProfileResourceModel struct {
	base.BaseResourceModel
	Type        types.String                        `tfsdk:"type"`
	EmailConfig *AlarmNotificationProfileEmailModel `tfsdk:"email_config"`
}

type AlarmNotificationProfileEmailModel struct {
	UseSMTPProfile types.Bool   `tfsdk:"use_smtp_profile"`
	EmailProfile   types.String `tfsdk:"email_profile"`
	Hostname       types.String `tfsdk:"hostname"`
	Port           types.Int64  `tfsdk:"port"`
	SSLEnabled     types.Bool   `tfsdk:"ssl_enabled"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
}

func (r *AlarmNotificationProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alarm_notification_profile"
}

func (r *AlarmNotificationProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Alarm Notification Profile in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the alarm notification profile.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the alarm notification profile.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alarm notification profile is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The type of the alarm notification profile (e.g., EmailNotificationProfileType).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"EmailNotificationProfileType",
					),
				},
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource, used for updates and deletes.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"email_config": schema.SingleNestedBlock{
				Description: "Configuration for Email Notification Profiles.",
				Attributes: map[string]schema.Attribute{
					"use_smtp_profile": schema.BoolAttribute{
						Description: "Whether to use an existing SMTP profile.",
						Required:    true,
					},
					"email_profile": schema.StringAttribute{
						Description: "The name of the SMTP profile to use (if use_smtp_profile is true).",
						Optional:    true,
					},
					"hostname": schema.StringAttribute{
						Description: "SMTP Server Hostname (if use_smtp_profile is false).",
						Optional:    true,
						Computed:    true,
					},
					"port": schema.Int64Attribute{
						Description: "SMTP Server Port (if use_smtp_profile is false).",
						Optional:    true,
						Computed:    true,
					},
					"ssl_enabled": schema.BoolAttribute{
						Description: "Enable SSL/TLS (if use_smtp_profile is false).",
						Optional:    true,
						Computed:    true,
					},
					"username": schema.StringAttribute{
						Description: "SMTP Username (if use_smtp_profile is false).",
						Optional:    true,
					},
					"password": schema.StringAttribute{
						Description: "SMTP Password (if use_smtp_profile is false).",
						Optional:    true,
						Sensitive:   true,
					},
				},
			},
		},
	}
}

func (r *AlarmNotificationProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.generic = base.GenericIgnitionResource[client.AlarmNotificationProfileConfig, AlarmNotificationProfileResourceModel]{
		Client:       apiClient,
		Handler:      r,
		Module:       "com.inductiveautomation.alarm-notification",
		ResourceType: "alarm-notification-profile",
		CreateFunc:   apiClient.CreateAlarmNotificationProfile,
		GetFunc:      apiClient.GetAlarmNotificationProfile,
		UpdateFunc:   apiClient.UpdateAlarmNotificationProfile,
		DeleteFunc:   apiClient.DeleteAlarmNotificationProfile,
	}
}

func (r *AlarmNotificationProfileResource) MapPlanToClient(ctx context.Context, model *AlarmNotificationProfileResourceModel) (client.AlarmNotificationProfileConfig, error) {
	config := client.AlarmNotificationProfileConfig{
		Profile: client.AlarmNotificationProfileProfile{
			Type: model.Type.ValueString(),
		},
		Settings: make(map[string]any),
	}

	if model.Type.ValueString() == "EmailNotificationProfileType" && model.EmailConfig != nil {
		emailSettings := make(map[string]any)
		emailSettings["useSmtpProfile"] = model.EmailConfig.UseSMTPProfile.ValueBool()

		if !model.EmailConfig.EmailProfile.IsNull() {
			emailSettings["emailProfile"] = model.EmailConfig.EmailProfile.ValueString()
		}
		if !model.EmailConfig.Hostname.IsNull() {
			emailSettings["hostname"] = model.EmailConfig.Hostname.ValueString()
		}
		if !model.EmailConfig.Port.IsNull() {
			emailSettings["port"] = int(model.EmailConfig.Port.ValueInt64())
		}
		if !model.EmailConfig.SSLEnabled.IsNull() {
			emailSettings["sslEnabled"] = model.EmailConfig.SSLEnabled.ValueBool()
		}
		if !model.EmailConfig.Username.IsNull() {
			emailSettings["username"] = model.EmailConfig.Username.ValueString()
		}
		if !model.EmailConfig.Password.IsNull() {
			encrypted, err := r.generic.Client.EncryptSecret(ctx, model.EmailConfig.Password.ValueString())
			if err != nil {
				return client.AlarmNotificationProfileConfig{}, err
			}
			emailSettings["password"] = encrypted
		}
		// In Ignition, these are nested under another "settings" key for this type
		config.Settings["settings"] = emailSettings
	}

	return config, nil
}

func (r *AlarmNotificationProfileResource) MapClientToState(ctx context.Context, name string, config *client.AlarmNotificationProfileConfig, model *AlarmNotificationProfileResourceModel) error {
	model.Name = types.StringValue(name)

	if config.Profile.Type == "EmailNotificationProfileType" {
		model.Type = types.StringValue(config.Profile.Type)

		if model.EmailConfig == nil {
			model.EmailConfig = &AlarmNotificationProfileEmailModel{}
		}

		if config.Settings != nil {
			settings := config.Settings
			if s, ok := settings["settings"].(map[string]any); ok {
				settings = s
			}

			if v, ok := settings["useSmtpProfile"].(bool); ok {
				model.EmailConfig.UseSMTPProfile = types.BoolValue(v)
			}

			emailProfile, _ := settings["emailProfile"].(string)
			model.EmailConfig.EmailProfile = base.GatewayString(model.EmailConfig.EmailProfile, emailProfile)

			hostname, _ := settings["hostname"].(string)
			model.EmailConfig.Hostname = base.GatewayString(model.EmailConfig.Hostname, hostname)

			if v, ok := settings["port"].(float64); ok && v != 0 {
				model.EmailConfig.Port = types.Int64Value(int64(v))
			} else if v, ok := settings["port"].(int); ok && v != 0 {
				model.EmailConfig.Port = types.Int64Value(int64(v))
			}

			if v, ok := settings["sslEnabled"].(bool); ok {
				model.EmailConfig.SSLEnabled = types.BoolValue(v)
			} else {
				model.EmailConfig.SSLEnabled = types.BoolValue(false)
			}

			username, _ := settings["username"].(string)
			model.EmailConfig.Username = base.GatewayString(model.EmailConfig.Username, username)
		} else if model.EmailConfig.SSLEnabled.IsUnknown() {
			model.EmailConfig.SSLEnabled = types.BoolValue(false)
		}
	} else {
		model.Type = base.GatewayString(model.Type, config.Profile.Type)
	}
	return nil
}

func (r *AlarmNotificationProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AlarmNotificationProfileResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AlarmNotificationProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AlarmNotificationProfileResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AlarmNotificationProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AlarmNotificationProfileResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AlarmNotificationProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AlarmNotificationProfileResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *AlarmNotificationProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &AlarmNotificationProfileResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
