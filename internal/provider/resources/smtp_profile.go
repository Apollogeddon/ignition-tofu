package resources

import (
	"context"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &SMTPProfileResource{}
	_ resource.ResourceWithImportState = &SMTPProfileResource{}
)

func NewSMTPProfileResource() resource.Resource {
	return &SMTPProfileResource{}
}

// SMTPProfileResource defines the resource implementation.
type SMTPProfileResource struct {
	client  client.IgnitionClient
	generic base.GenericIgnitionResource[client.SMTPProfileConfig, SMTPProfileResourceModel]
}

// SMTPProfileResourceModel describes the resource data model.
type SMTPProfileResourceModel struct {
	base.BaseResourceModel
	Hostname        types.String `tfsdk:"hostname"`
	Port            types.Int64  `tfsdk:"port"`
	UseSslPort      types.Bool   `tfsdk:"use_ssl_port"`
	StartTlsEnabled types.Bool   `tfsdk:"start_tls_enabled"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
}

func (r *SMTPProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp_profile"
}

func (r *SMTPProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an SMTP Email Profile in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the SMTP profile.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the SMTP profile.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the SMTP profile is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the SMTP server.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port SMTP service is running on.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(25),
			},
			"use_ssl_port": schema.BoolAttribute{
				Description: "Connect using dedicated SSL/TLS port.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"start_tls_enabled": schema.BoolAttribute{
				Description: "Connect using STARTTLS.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"username": schema.StringAttribute{
				Description: "The username for logging into the email server.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password for logging into the email server.",
				Optional:    true,
				Sensitive:   true,
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource, used for updates and deletes.",
				Computed:    true,
			},
		},
	}
}

func (r *SMTPProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.generic = base.GenericIgnitionResource[client.SMTPProfileConfig, SMTPProfileResourceModel]{
		Client:       c,
		Handler:      r,
		Module:       "ignition",
		ResourceType: "email-profile",
		CreateFunc:   c.CreateSMTPProfile,
		GetFunc:      c.GetSMTPProfile,
		UpdateFunc:   c.UpdateSMTPProfile,
		DeleteFunc:   c.DeleteSMTPProfile,
	}
}

func (r *SMTPProfileResource) MapPlanToClient(ctx context.Context, model *SMTPProfileResourceModel) (client.SMTPProfileConfig, error) {
	config := client.SMTPProfileConfig{
		Profile: client.SMTPProfileProfile{
			Type: "smtp.classic",
		},
		Settings: client.SMTPProfileSettings{
			Settings: &client.SMTPProfileSettingsClassic{
				Hostname:        model.Hostname.ValueString(),
				Port:            int(model.Port.ValueInt64()),
				UseSslPort:      model.UseSslPort.ValueBool(),
				StartTlsEnabled: model.StartTlsEnabled.ValueBool(),
			},
		},
	}

	if !model.Username.IsNull() {
		config.Settings.Settings.Username = model.Username.ValueString()
	}
	if !model.Password.IsNull() {
		encrypted, err := r.client.EncryptSecret(ctx, model.Password.ValueString())
		if err != nil {
			return client.SMTPProfileConfig{}, err
		}
		config.Settings.Settings.Password = encrypted
	}

	return config, nil
}

func (r *SMTPProfileResource) MapClientToState(ctx context.Context, name string, config *client.SMTPProfileConfig, model *SMTPProfileResourceModel) error {
	model.Name = types.StringValue(name)

	if config.Profile.Type != "" {
		model.Hostname = types.StringValue(config.Settings.Settings.Hostname)
		model.Port = types.Int64Value(int64(config.Settings.Settings.Port))
		model.UseSslPort = types.BoolValue(config.Settings.Settings.UseSslPort)
		model.StartTlsEnabled = types.BoolValue(config.Settings.Settings.StartTlsEnabled)
		model.Username = base.GatewayString(model.Username, config.Settings.Settings.Username)
	}
	return nil
}

func (r *SMTPProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SMTPProfileResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *SMTPProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SMTPProfileResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *SMTPProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SMTPProfileResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *SMTPProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SMTPProfileResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *SMTPProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &SMTPProfileResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
