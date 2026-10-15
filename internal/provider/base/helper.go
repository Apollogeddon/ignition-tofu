package base

import (
	"context"
	"errors"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BaseResourceModel includes the common fields for Ignition resources
type BaseResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	Signature   types.String `tfsdk:"signature"`
}

// IgnitionResourceHandler defines the unique mapping logic for a specific resource
type IgnitionResourceHandler[T any, M any] interface {
	// MapPlanToClient converts the Terraform model (plus plan data) to the API config struct
	MapPlanToClient(ctx context.Context, model *M) (T, error)
	// MapClientToState updates the Terraform model from the API response
	MapClientToState(ctx context.Context, name string, config *T, model *M) error
}

// GenericIgnitionResource implements the core CRUD logic
type GenericIgnitionResource[T any, M any] struct {
	Client       client.IgnitionClient
	Handler      IgnitionResourceHandler[T, M]
	Module       string
	ResourceType string

	CreateFunc func(context.Context, client.ResourceResponse[T]) (*client.ResourceResponse[T], error)
	GetFunc    func(context.Context, string) (*client.ResourceResponse[T], error)
	UpdateFunc func(context.Context, client.ResourceResponse[T]) (*client.ResourceResponse[T], error)
	DeleteFunc func(context.Context, string, string) error
}

func (r *GenericIgnitionResource[T, M]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, data *M, baseModel *BaseResourceModel) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.Handler.MapPlanToClient(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping plan to client", err.Error())
		return
	}

	res := client.ResourceResponse[T]{
		Module:  r.Module,
		Type:    r.ResourceType,
		Name:    baseModel.Name.ValueString(),
		Enabled: BoolPtr(baseModel.Enabled.ValueBool()),
		Config:  config,
	}

	if !baseModel.Description.IsNull() {
		res.Description = baseModel.Description.ValueString()
	}

	created, err := r.CreateFunc(ctx, res)
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource", err.Error())
		return
	}

	baseModel.Signature = types.StringValue(created.Signature)
	baseModel.Id = types.StringValue(created.Name)
	if baseModel.Name.IsNull() || baseModel.Name.IsUnknown() || baseModel.Name.ValueString() == "" {
		baseModel.Name = types.StringValue(created.Name)
	}
	if created.Enabled != nil {
		baseModel.Enabled = types.BoolValue(*created.Enabled)
	} else {
		baseModel.Enabled = types.BoolValue(true)
	}
	if created.Description != "" {
		baseModel.Description = types.StringValue(created.Description)
	} else if baseModel.Description.IsNull() || baseModel.Description.IsUnknown() {
		baseModel.Description = types.StringNull()
	}

	if err := r.Handler.MapClientToState(ctx, created.Name, &created.Config, data); err != nil {
		resp.Diagnostics.AddError("Error mapping client to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *GenericIgnitionResource[T, M]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, data *M, baseModel *BaseResourceModel) {
	resp.Diagnostics.Append(req.State.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if baseModel.Name.ValueString() == "" {
		return
	}

	res, err := r.GetFunc(ctx, baseModel.Name.ValueString())
	if err != nil {
		HandleReadError(ctx, resp, err)
		return
	}

	baseModel.Signature = types.StringValue(res.Signature)
	baseModel.Id = types.StringValue(res.Name)
	if baseModel.Name.IsNull() || baseModel.Name.IsUnknown() || baseModel.Name.ValueString() == "" {
		baseModel.Name = types.StringValue(res.Name)
	}
	if res.Enabled != nil {
		baseModel.Enabled = types.BoolValue(*res.Enabled)
	} else {
		baseModel.Enabled = types.BoolValue(true)
	}

	baseModel.Description = GatewayString(baseModel.Description, res.Description)

	if err := r.Handler.MapClientToState(ctx, res.Name, &res.Config, data); err != nil {
		resp.Diagnostics.AddError("Error mapping client to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// HandleReadError reports a failed lookup in a resource's Read. A resource the
// gateway no longer has (deleted outside OpenTofu) is removed from state, so
// the next plan recreates it rather than failing.
func HandleReadError(ctx context.Context, resp *resource.ReadResponse, err error) {
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.AddError("Error reading resource", err.Error())
}

func (r *GenericIgnitionResource[T, M]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, data *M, baseModel *BaseResourceModel) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve existing signature from state
	var sig types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("signature"), &sig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.Handler.MapPlanToClient(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Error mapping plan to client", err.Error())
		return
	}

	res := client.ResourceResponse[T]{
		Module:    r.Module,
		Type:      r.ResourceType,
		Name:      baseModel.Name.ValueString(),
		Enabled:   BoolPtr(baseModel.Enabled.ValueBool()),
		Signature: sig.ValueString(),
		Config:    config,
	}

	if !baseModel.Description.IsNull() {
		res.Description = baseModel.Description.ValueString()
	}

	updated, err := r.UpdateFunc(ctx, res)
	if err != nil {
		resp.Diagnostics.AddError("Error updating resource", err.Error())
		return
	}

	if updated.Signature != "" {
		baseModel.Signature = types.StringValue(updated.Signature)
	} else {
		// Attempt to fetch the latest signature if not returned by Update
		fresh, err := r.GetFunc(ctx, baseModel.Name.ValueString())
		if err == nil && fresh.Signature != "" {
			baseModel.Signature = types.StringValue(fresh.Signature)
			updated = fresh
		} else if !sig.IsNull() && !sig.IsUnknown() {
			baseModel.Signature = sig
			resp.Diagnostics.AddWarning("Missing Signature on Update",
				"The API returned an empty signature after update and refresh failed. Preserving the existing signature.")
		}
	}

	baseModel.Id = types.StringValue(updated.Name)
	if baseModel.Name.IsNull() || baseModel.Name.IsUnknown() || baseModel.Name.ValueString() == "" {
		baseModel.Name = types.StringValue(updated.Name)
	}

	if updated.Enabled != nil {
		baseModel.Enabled = types.BoolValue(*updated.Enabled)
	} else {
		baseModel.Enabled = types.BoolValue(true)
	}
	if updated.Description != "" {
		baseModel.Description = types.StringValue(updated.Description)
	} else if baseModel.Description.IsNull() || baseModel.Description.IsUnknown() {
		baseModel.Description = types.StringNull()
	}

	if err := r.Handler.MapClientToState(ctx, updated.Name, &updated.Config, data); err != nil {
		resp.Diagnostics.AddError("Error mapping client to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *GenericIgnitionResource[T, M]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, data *M, baseModel *BaseResourceModel) {
	resp.Diagnostics.Append(req.State.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A resource already deleted outside OpenTofu is already where destroy wants it
	err := r.DeleteFunc(ctx, baseModel.Name.ValueString(), baseModel.Signature.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting resource", err.Error())
		return
	}
}

// GatewayString maps a string the gateway returned onto the value in state. A
// value cleared on the gateway nulls a non-empty one so plan shows it as drift,
// while an empty string the user set stays empty rather than diffing forever.
func GatewayString(current types.String, s string) types.String {
	if s != "" {
		return types.StringValue(s)
	}
	if current.ValueString() != "" || current.IsUnknown() {
		return types.StringNull()
	}
	return current
}

func BoolPtr(b bool) *bool {
	return &b
}

// KnownBool is v's value when the plan knows it, else nil, which leaves the gateway's value.
func KnownBool(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return BoolPtr(v.ValueBool())
}

// KnownInt is v's value when the plan knows it, else nil, which leaves the gateway's value.
func KnownInt(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	n := int(v.ValueInt64())
	return &n
}

// Deref is *p, or the zero value when the gateway sent nothing.
func Deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
