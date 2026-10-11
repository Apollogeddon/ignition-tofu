package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &IdentityProviderResource{}
	_ resource.ResourceWithImportState = &IdentityProviderResource{}
)

func NewIdentityProviderResource() resource.Resource {
	return &IdentityProviderResource{}
}

// IdentityProviderResource defines the resource implementation.
type IdentityProviderResource struct {
	client  client.IgnitionClient
	generic base.GenericIgnitionResource[client.IdentityProviderConfig, IdentityProviderResourceModel]
}

// IdentityProviderResourceModel describes the resource data model.
type IdentityProviderResourceModel struct {
	base.BaseResourceModel
	Type                     types.String  `tfsdk:"type"`
	UserSource               types.String  `tfsdk:"user_source"`
	SessionInactivityTimeout types.Float64 `tfsdk:"session_inactivity_timeout"`
	SessionExp               types.Float64 `tfsdk:"session_expiration"`
	RememberMeExp            types.Float64 `tfsdk:"remember_me_expiration"`
	ClientId                 types.String  `tfsdk:"client_id"`
	ClientSecret             types.String  `tfsdk:"client_secret"`
	ProviderId               types.String  `tfsdk:"provider_id"`
	AuthorizationEndpoint    types.String  `tfsdk:"authorization_endpoint"`
	TokenEndpoint            types.String  `tfsdk:"token_endpoint"`
	JwkEndpoint              types.String  `tfsdk:"jwk_endpoint"`
	JwkEndpointEnabled       types.Bool    `tfsdk:"jwk_endpoint_enabled"`
	UserInfoEndpoint         types.String  `tfsdk:"user_info_endpoint"`
	LogoutEndpoint           types.String  `tfsdk:"logout_endpoint"`
	// SAML fields
	IdpEntityId                 types.String      `tfsdk:"idp_entity_id"`
	SpEntityId                  types.String      `tfsdk:"sp_entity_id"`
	SpEntityIdEnabled           types.Bool        `tfsdk:"sp_entity_id_enabled"`
	AcsBinding                  types.String      `tfsdk:"acs_binding"`
	NameIdFormat                types.String      `tfsdk:"name_id_format"`
	SsoServiceConfig            *SsoServiceConfig `tfsdk:"sso_service_config"`
	ForceAuthn                  types.Bool        `tfsdk:"force_authn"`
	ResponseSignaturesRequired  types.Bool        `tfsdk:"response_signatures_required"`
	AssertionSignaturesRequired types.Bool        `tfsdk:"assertion_signatures_required"`
	IdpMetadataUrl              types.String      `tfsdk:"idp_metadata_url"`
	IdpMetadataUrlEnabled       types.Bool        `tfsdk:"idp_metadata_url_enabled"`
}

type SsoServiceConfig struct {
	Uri     types.String `tfsdk:"uri"`
	Binding types.String `tfsdk:"binding"`
}

func (r *IdentityProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_provider"
}

func (r *IdentityProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Identity Provider in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the identity provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the identity provider.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the identity provider is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The type of the identity provider (internal, oidc, saml).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("internal", "oidc", "saml"),
				},
			},
			// Internal Type Fields
			"user_source": schema.StringAttribute{
				Description: "The name of the User Source Profile used to authenticate users (for 'internal' type).",
				Optional:    true,
			},
			"session_inactivity_timeout": schema.Float64Attribute{
				Description: "Minutes before expiring a session due to user inactivity.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(0),
			},
			"session_expiration": schema.Float64Attribute{
				Description: "Maximum minutes a session may exist before it is expired.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(0),
			},
			"remember_me_expiration": schema.Float64Attribute{
				Description: "Maximum hours a user will be remembered.",
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(0),
			},
			// OIDC Type Fields
			"client_id": schema.StringAttribute{
				Description: "The client identifier registered within the identity provider.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "The client secret registered within the identity provider.",
				Optional:    true,
				Sensitive:   true,
			},
			"provider_id": schema.StringAttribute{
				Description: "The issuer URL of the identity provider.",
				Optional:    true,
			},
			"authorization_endpoint": schema.StringAttribute{
				Description: "URL of the OP's OAuth 2.0 Authorization Endpoint.",
				Optional:    true,
			},
			"token_endpoint": schema.StringAttribute{
				Description: "URL of the OP's OAuth 2.0 Token Endpoint.",
				Optional:    true,
			},
			"jwk_endpoint": schema.StringAttribute{
				Description: "URL of the OP's JSON Web Key Set document.",
				Optional:    true,
			},
			"jwk_endpoint_enabled": schema.BoolAttribute{
				Description: "If true, then identity provider public keys will be automatically downloaded.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"user_info_endpoint": schema.StringAttribute{
				Description: "URL to retrieve UserInfo claims from the provider.",
				Optional:    true,
			},
			"logout_endpoint": schema.StringAttribute{
				Description: "URL at the OP to which an RP can perform a redirect to request that the End-User be logged out.",
				Optional:    true,
			},
			// SAML Type Fields
			"idp_entity_id": schema.StringAttribute{
				Description: "The Identity Provider's Entity ID.",
				Optional:    true,
			},
			"sp_entity_id": schema.StringAttribute{
				Description: "The Service Provider's Entity ID.",
				Optional:    true,
			},
			"sp_entity_id_enabled": schema.BoolAttribute{
				Description: "True if the SP Entity ID setting should be used.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"acs_binding": schema.StringAttribute{
				Description: "The expected binding used by the IdP (HTTP-Redirect, HTTP-POST).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"),
			},
			"name_id_format": schema.StringAttribute{
				Description: "The expected name ID format.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"),
			},
			"sso_service_config": schema.SingleNestedAttribute{
				Description: "The Identity Provider's SSO Service Configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"uri":     schema.StringAttribute{Description: "The URL of the IdP's single sign-on service.", Required: true},
					"binding": schema.StringAttribute{Description: "The SAML binding of the single sign-on service, e.g. urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect.", Required: true},
				},
			},
			"force_authn": schema.BoolAttribute{
				Description: "Whether to ask the IdP to authenticate the user again rather than reuse an existing session (SAML ForceAuthn).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"response_signatures_required": schema.BoolAttribute{
				Description: "Whether SAML responses must be signed.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"assertion_signatures_required": schema.BoolAttribute{
				Description: "Whether SAML assertions must be signed.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"idp_metadata_url": schema.StringAttribute{
				Description: "The URL of the IdP's SAML metadata.",
				Optional:    true,
			},
			"idp_metadata_url_enabled": schema.BoolAttribute{
				Description: "True if the IdP metadata URL setting should be used.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource.",
				Computed:    true,
			},
		},
	}
}

func (r *IdentityProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.generic = base.GenericIgnitionResource[client.IdentityProviderConfig, IdentityProviderResourceModel]{
		Client:       c,
		Handler:      r,
		Module:       "ignition",
		ResourceType: "identity-provider",
		CreateFunc:   c.CreateIdentityProvider,
		GetFunc:      c.GetIdentityProvider,
		UpdateFunc:   c.UpdateIdentityProvider,
		DeleteFunc:   c.DeleteIdentityProvider,
	}
}

// defaultIdentityProviderProfile builds a profile with the gateway's required
// (but currently unconfigurable via this resource) attribute-mapper, grant,
// and security-level-rule fields left at sensible defaults: id and userName
// are mapped directly to the user source's username attribute, matching
// what the Designer produces for a basic identity provider.
func defaultIdentityProviderProfile(idpType string) client.IdentityProviderProfile {
	directUsernameMapper := map[string]any{
		"type":   "direct",
		"config": map[string]any{"attributePath": "username"},
	}
	return client.IdentityProviderProfile{
		Type: idpType,
		UserAttributeMapper: map[string]any{
			"id":       directUsernameMapper,
			"userName": directUsernameMapper,
		},
		UserGrants:         map[string]any{},
		SecurityLevelRules: client.IdentityProviderSecurityLevelRules{Nodes: []any{}},
	}
}

func (r *IdentityProviderResource) MapPlanToClient(ctx context.Context, model *IdentityProviderResourceModel) (client.IdentityProviderConfig, error) {
	switch model.Type.ValueString() {
	case "internal":
		internalConfig := client.IdentityProviderInternalConfig{
			UserSource:               model.UserSource.ValueString(),
			SessionInactivityTimeout: model.SessionInactivityTimeout.ValueFloat64(),
			SessionExp:               model.SessionExp.ValueFloat64(),
			RememberMeExp:            model.RememberMeExp.ValueFloat64(),
			AuthMethods: []client.IdentityProviderAuthMethod{
				{
					Type:   "basic",
					Config: map[string]any{},
				},
			},
		}
		return client.IdentityProviderConfig{
			Profile:  defaultIdentityProviderProfile("internal"),
			Settings: internalConfig,
		}, nil
	case "oidc":
		oidcConfig := client.IdentityProviderOidcConfig{
			ClientId:                   model.ClientId.ValueString(),
			ProviderId:                 model.ProviderId.ValueString(),
			AuthorizationEndpoint:      model.AuthorizationEndpoint.ValueString(),
			TokenEndpoint:              model.TokenEndpoint.ValueString(),
			JsonWebKeysEndpoint:        model.JwkEndpoint.ValueString(),
			JsonWebKeysEndpointEnabled: model.JwkEndpointEnabled.ValueBool(),
			UserInfoEndpoint:           model.UserInfoEndpoint.ValueString(),
			EndSessionEndpoint:         model.LogoutEndpoint.ValueString(),
		}

		if !model.ClientSecret.IsNull() {
			encrypted, err := r.client.EncryptSecret(ctx, model.ClientSecret.ValueString())
			if err != nil {
				return client.IdentityProviderConfig{}, err
			}
			oidcConfig.ClientSecret = encrypted
		}

		return client.IdentityProviderConfig{
			Profile:  defaultIdentityProviderProfile("oidc"),
			Settings: oidcConfig,
		}, nil
	case "saml":
		samlConfig := client.IdentityProviderSamlConfig{
			IdpEntityId:                    model.IdpEntityId.ValueString(),
			SpEntityId:                     model.SpEntityId.ValueString(),
			SpEntityIdEnabled:              model.SpEntityIdEnabled.ValueBool(),
			AcsBinding:                     model.AcsBinding.ValueString(),
			NameIdFormat:                   model.NameIdFormat.ValueString(),
			ForceAuthnEnabled:              model.ForceAuthn.ValueBool(),
			ResponseSignaturesRequired:     model.ResponseSignaturesRequired.ValueBool(),
			AssertionSignaturesRequired:    model.AssertionSignaturesRequired.ValueBool(),
			IdpMetadataUrl:                 model.IdpMetadataUrl.ValueString(),
			IdpMetadataUrlEnabled:          model.IdpMetadataUrlEnabled.ValueBool(),
			SignatureVerifyingCertificates: []string{},
			SignatureVerifyingKeys:         []any{},
		}

		if model.SsoServiceConfig != nil {
			samlConfig.SsoServiceConfig.Uri = model.SsoServiceConfig.Uri.ValueString()
			samlConfig.SsoServiceConfig.Binding = model.SsoServiceConfig.Binding.ValueString()
		}

		return client.IdentityProviderConfig{
			Profile:  defaultIdentityProviderProfile("saml"),
			Settings: samlConfig,
		}, nil
	}

	return client.IdentityProviderConfig{}, fmt.Errorf("unsupported identity provider type: %s", model.Type.ValueString())
}

func (r *IdentityProviderResource) MapClientToState(ctx context.Context, name string, config *client.IdentityProviderConfig, model *IdentityProviderResourceModel) error {
	model.Name = types.StringValue(name)
	model.Type = types.StringValue(config.Profile.Type)

	// Seed every Computed+Default attribute with its schema default so that
	// Read/Import produce the same state as Create for whichever type-specific
	// fields don't apply to this identity provider's type. The per-type cases
	// below overwrite the ones that are actually relevant.
	model.SessionInactivityTimeout = types.Float64Value(0)
	model.SessionExp = types.Float64Value(0)
	model.RememberMeExp = types.Float64Value(0)
	model.JwkEndpointEnabled = types.BoolValue(true)
	model.SpEntityIdEnabled = types.BoolValue(false)
	model.AcsBinding = types.StringValue("urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST")
	model.NameIdFormat = types.StringValue("urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified")
	model.ForceAuthn = types.BoolValue(false)
	model.ResponseSignaturesRequired = types.BoolValue(true)
	model.AssertionSignaturesRequired = types.BoolValue(true)
	model.IdpMetadataUrlEnabled = types.BoolValue(true)

	configBytes, _ := json.Marshal(config.Settings)

	switch config.Profile.Type {
	case "internal":
		var internalConfig client.IdentityProviderInternalConfig
		if err := json.Unmarshal(configBytes, &internalConfig); err == nil {
			model.UserSource = types.StringValue(internalConfig.UserSource)
			model.SessionInactivityTimeout = types.Float64Value(internalConfig.SessionInactivityTimeout)
			model.SessionExp = types.Float64Value(internalConfig.SessionExp)
			model.RememberMeExp = types.Float64Value(internalConfig.RememberMeExp)
		}
	case "oidc":
		var oidcConfig client.IdentityProviderOidcConfig
		if err := json.Unmarshal(configBytes, &oidcConfig); err == nil {
			model.ClientId = types.StringValue(oidcConfig.ClientId)
			model.ProviderId = types.StringValue(oidcConfig.ProviderId)
			model.AuthorizationEndpoint = base.StringToNullableString(oidcConfig.AuthorizationEndpoint)
			model.TokenEndpoint = base.StringToNullableString(oidcConfig.TokenEndpoint)
			model.JwkEndpoint = base.StringToNullableString(oidcConfig.JsonWebKeysEndpoint)
			model.JwkEndpointEnabled = types.BoolValue(oidcConfig.JsonWebKeysEndpointEnabled)
			model.UserInfoEndpoint = base.StringToNullableString(oidcConfig.UserInfoEndpoint)
			model.LogoutEndpoint = base.StringToNullableString(oidcConfig.EndSessionEndpoint)
		}
	case "saml":
		var samlConfig client.IdentityProviderSamlConfig
		if err := json.Unmarshal(configBytes, &samlConfig); err == nil {
			model.IdpEntityId = types.StringValue(samlConfig.IdpEntityId)
			model.SpEntityId = base.StringToNullableString(samlConfig.SpEntityId)
			model.SpEntityIdEnabled = types.BoolValue(samlConfig.SpEntityIdEnabled)
			model.AcsBinding = types.StringValue(samlConfig.AcsBinding)
			model.NameIdFormat = types.StringValue(samlConfig.NameIdFormat)
			model.ForceAuthn = types.BoolValue(samlConfig.ForceAuthnEnabled)
			model.ResponseSignaturesRequired = types.BoolValue(samlConfig.ResponseSignaturesRequired)
			model.AssertionSignaturesRequired = types.BoolValue(samlConfig.AssertionSignaturesRequired)
			model.IdpMetadataUrl = base.StringToNullableString(samlConfig.IdpMetadataUrl)
			model.IdpMetadataUrlEnabled = types.BoolValue(samlConfig.IdpMetadataUrlEnabled)

			model.SsoServiceConfig = &SsoServiceConfig{
				Uri:     types.StringValue(samlConfig.SsoServiceConfig.Uri),
				Binding: types.StringValue(samlConfig.SsoServiceConfig.Binding),
			}
		}
	}

	return nil
}

func (r *IdentityProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IdentityProviderResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *IdentityProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IdentityProviderResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *IdentityProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IdentityProviderResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *IdentityProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IdentityProviderResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *IdentityProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &IdentityProviderResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
