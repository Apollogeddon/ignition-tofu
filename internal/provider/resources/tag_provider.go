package resources

import (
	"context"
	"fmt"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &TagProviderResource{}
	_ resource.ResourceWithImportState = &TagProviderResource{}
)

func NewTagProviderResource() resource.Resource {
	return &TagProviderResource{}
}

// TagProviderResource defines the resource implementation.
type TagProviderResource struct {
	generic base.GenericIgnitionResource[client.TagProviderConfig, TagProviderResourceModel]
}

// TagProviderResourceModel describes the resource data model.
type TagProviderResourceModel struct {
	base.BaseResourceModel
	Type types.String `tfsdk:"type"`
}

func (r *TagProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag_provider"
}

func (r *TagProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Tag Provider in Ignition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the tag provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of the tag provider (e.g., standard).",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the tag provider.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the tag provider is enabled.",
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

func (r *TagProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.generic = base.GenericIgnitionResource[client.TagProviderConfig, TagProviderResourceModel]{
		Client:       c,
		Handler:      r,
		Module:       "ignition",
		ResourceType: "tag-provider",
		CreateFunc:   c.CreateTagProvider,
		GetFunc:      c.GetTagProvider,
		UpdateFunc:   c.UpdateTagProvider,
		DeleteFunc:   c.DeleteTagProvider,
	}
}

func (r *TagProviderResource) MapPlanToClient(ctx context.Context, model *TagProviderResourceModel) (client.TagProviderConfig, error) {
	return client.TagProviderConfig{
		Profile: client.TagProviderProfile{
			Type: model.Type.ValueString(),
		},
		Description: model.Description.ValueString(),
		Settings:    make(map[string]any),
	}, nil
}

func (r *TagProviderResource) MapClientToState(ctx context.Context, name string, config *client.TagProviderConfig, model *TagProviderResourceModel) error {
	model.Name = types.StringValue(name)
	model.Type = types.StringValue(config.Profile.Type)
	model.Description = base.GatewayString(model.Description, config.Description)
	return nil
}

func (r *TagProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TagProviderResourceModel
	r.generic.Create(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *TagProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TagProviderResourceModel
	r.generic.Read(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *TagProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TagProviderResourceModel
	r.generic.Update(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *TagProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TagProviderResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *TagProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &TagProviderResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
