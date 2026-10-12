package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Numbers are checked at plan time: a port outside 1-65535, or a rate of 0 the
// gateway would run continuously, never reaches it.
func TestNumericAttributesAreRangeChecked(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		res       resource.Resource
		at        path.Path
		bad, good int64
	}{
		{NewGanOutgoingResource(), path.Root("port"), 0, 8060},
		{NewGanOutgoingResource(), path.Root("port"), 65536, 8060},
		{NewSMTPProfileResource(), path.Root("port"), -1, 25},
		{NewAlarmNotificationProfileResource(), path.Root("email_config").AtName("port"), 70000, 587},
		{NewRedundancyResource(), path.Root("gateway_network_setup").AtName("port"), 0, 8060},
		{NewRedundancyResource(), path.Root("join_wait_time"), -1, 0},
		{NewAuditProfileResource(), path.Root("retention_days"), -1, 0},
		{NewStoreAndForwardResource(), path.Root("time_threshold_ms"), 0, 1000},
		{NewStoreAndForwardResource(), path.Root("forward_rate_ms"), 0, 1000},
		{NewStoreAndForwardResource(), path.Root("scan_rate_ms"), 0, 1000},
		{NewStoreAndForwardResource(), path.Root("batch_size"), 0, 100},
		{NewStoreAndForwardResource(), path.Root("data_threshold"), 0, 100},
		{NewStoreAndForwardResource(), path.Root("primary_policy").AtName("value"), -1, 0},
	}
	for _, tc := range cases {
		var resp resource.SchemaResponse
		tc.res.Schema(ctx, resource.SchemaRequest{}, &resp)
		attr, diags := resp.Schema.AttributeAtPath(ctx, tc.at)
		if diags.HasError() {
			t.Fatalf("%s: %v", tc.at, diags)
		}
		check := func(v int64) bool {
			var out validator.Int64Response
			for _, val := range attr.(schema.Int64Attribute).Validators {
				val.ValidateInt64(ctx, validator.Int64Request{Path: tc.at, ConfigValue: types.Int64Value(v)}, &out)
			}
			return out.Diagnostics.HasError()
		}
		if !check(tc.bad) {
			t.Errorf("%s: %d accepted", tc.at, tc.bad)
		}
		if check(tc.good) {
			t.Errorf("%s: %d refused", tc.at, tc.good)
		}
	}
}
