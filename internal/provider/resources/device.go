package resources

import (
	"context"
	"encoding/json"
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

var (
	_ resource.Resource                = &DeviceResource{}
	_ resource.ResourceWithImportState = &DeviceResource{}
)

func NewDeviceResource() resource.Resource {
	return &DeviceResource{}
}

type DeviceResource struct {
	generic base.GenericIgnitionResource[client.DeviceConfig, DeviceResourceModel]
}

type DeviceResourceModel struct {
	base.BaseResourceModel
	Type       types.String `tfsdk:"type"`
	Parameters types.String `tfsdk:"parameters"`
}

func (r *DeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *DeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Device (Driver) in Ignition (e.g., Modbus, Siemens, Simulator).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the device.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the device.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the device is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The type of the device (e.g., 'ModbusTcp', 'S71500', 'ProgrammableSimulatorDevice').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parameters": schema.StringAttribute{
				Description: "The JSON configuration parameters for the device. These vary by device type.",
				Required:    true,
			},
			"signature": schema.StringAttribute{
				Description: "The signature of the resource, used for updates and deletes.",
				Computed:    true,
			},
		},
	}
}

func (r *DeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.IgnitionClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.IgnitionClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.generic.Client = client
	r.generic.Handler = r
	r.generic.Module = "com.inductiveautomation.opcua"
	r.generic.ResourceType = "device"
	r.generic.CreateFunc = client.CreateDevice
	r.generic.GetFunc = client.GetDevice
	r.generic.UpdateFunc = client.UpdateDevice
	r.generic.DeleteFunc = client.DeleteDevice
}

func (r *DeviceResource) MapPlanToClient(ctx context.Context, model *DeviceResourceModel) (client.DeviceConfig, error) {
	config := make(client.DeviceConfig)

	// Unmarshal JSON parameters
	if err := json.Unmarshal([]byte(model.Parameters.ValueString()), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters JSON: %w", err)
	}

	return config, nil
}

func (r *DeviceResource) MapClientToState(ctx context.Context, name string, config *client.DeviceConfig, model *DeviceResourceModel) error {
	model.Name = types.StringValue(name)

	// Marshal the API response to canonical JSON
	b, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters to JSON: %w", err)
	}
	canonicalAPI := string(b)

	if !model.Parameters.IsNull() && !model.Parameters.IsUnknown() {
		var modelConfig client.DeviceConfig
		if err := json.Unmarshal([]byte(model.Parameters.ValueString()), &modelConfig); err == nil {
			bModel, _ := json.Marshal(modelConfig)
			if string(bModel) == canonicalAPI {
				// They are semantically equal (ignoring whitespace), so keep the Plan value
				// to satisfy Terraform's consistency check.
				return nil
			}
		}
	}

	// If they differ (or model is null/unknown), accept the API's canonical value
	model.Parameters = types.StringValue(canonicalAPI)

	return nil
}

func (r *DeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeviceResourceModel
	if r.generic.Client == nil {
		base.AddUnconfiguredError(&resp.Diagnostics)
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.MapPlanToClient(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping plan to client", err.Error())
		return
	}

	res := client.ResourceResponse[client.DeviceConfig]{
		Module:  r.generic.Module,
		Type:    data.Type.ValueString(), // Use the driver type from the plan
		Name:    data.Name.ValueString(),
		Enabled: base.BoolPtr(data.Enabled.ValueBool()),
		Config:  config,
	}

	if !data.Description.IsNull() {
		res.Description = data.Description.ValueString()
	}

	created, err := r.generic.CreateFunc(ctx, res)
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource", err.Error())
		return
	}

	data.Signature = types.StringValue(created.Signature)
	data.Id = types.StringValue(created.Name)
	if data.Name.IsNull() || data.Name.IsUnknown() || data.Name.ValueString() == "" {
		data.Name = types.StringValue(created.Name)
	}
	if created.Enabled != nil {
		data.Enabled = types.BoolValue(*created.Enabled)
	} else {
		data.Enabled = types.BoolValue(true)
	}
	if created.Description != "" {
		data.Description = types.StringValue(created.Description)
	} else if data.Description.IsNull() || data.Description.IsUnknown() {
		data.Description = types.StringNull()
	}

	// The API returns the Driver Type in the Type field, so we map it back
	data.Type = types.StringValue(created.Type)

	if err := r.MapClientToState(ctx, created.Name, &created.Config, &data); err != nil {
		resp.Diagnostics.AddError("Error mapping client to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *DeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Name.ValueString() == "" || base.Unconfigured(r.generic.Client, &resp.Diagnostics) {
		return
	}

	res, err := r.generic.GetFunc(ctx, data.Name.ValueString())
	if err != nil {
		base.HandleReadError(ctx, resp, err)
		return
	}

	data.Signature = types.StringValue(res.Signature)
	data.Id = types.StringValue(res.Name)
	if data.Name.IsNull() || data.Name.IsUnknown() || data.Name.ValueString() == "" {
		data.Name = types.StringValue(res.Name)
	}
	if res.Enabled != nil {
		data.Enabled = types.BoolValue(*res.Enabled)
	} else {
		data.Enabled = types.BoolValue(true)
	}
	if res.Description != "" {
		data.Description = types.StringValue(res.Description)
	} else if data.Description.IsNull() || data.Description.IsUnknown() {
		data.Description = types.StringNull()
	}
	// The driver type lives on the resource's own type field, not inside
	// config.profile.type as consumed by MapClientToState.
	data.Type = types.StringValue(res.Type)

	if err := r.MapClientToState(ctx, res.Name, &res.Config, &data); err != nil {
		resp.Diagnostics.AddError("Error mapping client to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *DeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeviceResourceModel
	if r.generic.Client == nil {
		base.AddUnconfiguredError(&resp.Diagnostics)
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve existing signature from state
	var stateModel DeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.MapPlanToClient(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping plan to client", err.Error())
		return
	}

	res := client.ResourceResponse[client.DeviceConfig]{
		Module:    r.generic.Module,
		Type:      data.Type.ValueString(), // Use the driver type from the plan
		Name:      data.Name.ValueString(),
		Enabled:   base.BoolPtr(data.Enabled.ValueBool()),
		Signature: stateModel.Signature.ValueString(),
		Config:    config,
	}

	if !data.Description.IsNull() {
		res.Description = data.Description.ValueString()
	}

	updated, err := r.generic.UpdateFunc(ctx, res)
	if err != nil {
		resp.Diagnostics.AddError("Error updating resource", err.Error())
		return
	}

	if updated.Signature != "" {
		data.Signature = types.StringValue(updated.Signature)
	} else {
		fresh, err := r.generic.GetFunc(ctx, data.Name.ValueString())
		if err == nil && fresh.Signature != "" {
			data.Signature = types.StringValue(fresh.Signature)
			updated = fresh
		} else if !stateModel.Signature.IsNull() && !stateModel.Signature.IsUnknown() {
			data.Signature = stateModel.Signature
			resp.Diagnostics.AddWarning("Missing Signature on Update",
				"The API returned an empty signature after update and refresh failed. Preserving the existing signature.")
		}
	}

	data.Id = types.StringValue(updated.Name)
	if data.Name.IsNull() || data.Name.IsUnknown() || data.Name.ValueString() == "" {
		data.Name = types.StringValue(updated.Name)
	}

	if updated.Enabled != nil {
		data.Enabled = types.BoolValue(*updated.Enabled)
	} else {
		data.Enabled = types.BoolValue(true)
	}
	if updated.Description != "" {
		data.Description = types.StringValue(updated.Description)
	} else if data.Description.IsNull() || data.Description.IsUnknown() {
		data.Description = types.StringNull()
	}

	// Map the Type back
	data.Type = types.StringValue(updated.Type)

	if err := r.MapClientToState(ctx, updated.Name, &updated.Config, &data); err != nil {
		resp.Diagnostics.AddError("Error mapping client to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *DeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeviceResourceModel
	r.generic.Delete(ctx, req, resp, &data, &data.BaseResourceModel)
}

func (r *DeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &DeviceResourceModel{
		BaseResourceModel: base.BaseResourceModel{
			Id:   types.StringValue(req.ID),
			Name: types.StringValue(req.ID),
		},
	})...)
}
