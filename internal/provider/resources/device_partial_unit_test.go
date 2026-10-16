package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// partialDeviceGateway stores a device's settings as the gateway does: every
// setting the request leaves out takes the driver's default.
func partialDeviceGateway() (*client.MockClient, func(func(client.DeviceConfig))) {
	defaults := client.DeviceConfig{
		"hostname":    "",
		"port":        float64(502),
		"connectTime": float64(10000),
		"addressing":  map[string]any{"zeroBased": false, "maxHoldingRegisters": float64(125)},
	}
	var stored client.DeviceConfig
	store := func(item client.ResourceResponse[client.DeviceConfig]) *client.ResourceResponse[client.DeviceConfig] {
		stored = maps.Clone(defaults)
		for k, v := range item.Config {
			if nested, ok := v.(map[string]any); ok {
				merged := maps.Clone(defaults[k].(map[string]any))
				maps.Copy(merged, nested)
				v = merged
			}
			stored[k] = v
		}
		item.Config = maps.Clone(stored)
		item.Signature = "sig"
		return &item
	}
	mock := &client.MockClient{
		CreateDeviceFunc: func(_ context.Context, item client.ResourceResponse[client.DeviceConfig]) (*client.ResourceResponse[client.DeviceConfig], error) {
			return store(item), nil
		},
		UpdateDeviceFunc: func(_ context.Context, item client.ResourceResponse[client.DeviceConfig]) (*client.ResourceResponse[client.DeviceConfig], error) {
			return store(item), nil
		},
		GetDeviceFunc: func(_ context.Context, name string) (*client.ResourceResponse[client.DeviceConfig], error) {
			return &client.ResourceResponse[client.DeviceConfig]{
				Name:      name,
				Type:      "ModbusTcp",
				Enabled:   base.BoolPtr(true),
				Signature: "sig",
				Config:    maps.Clone(stored),
			}, nil
		},
		DeleteDeviceFunc: func(context.Context, string, string) error { return nil },
	}
	return mock, func(change func(client.DeviceConfig)) { change(stored) }
}

func partialDeviceConfig(parameters string) string {
	return fmt.Sprintf(`
provider "ignition" {
  host  = "http://mock-host"
  token = "mock-token"
}
resource "ignition_device" "test" {
  name       = "Modbus"
  type       = "ModbusTcp"
  parameters = %q
}
`, parameters)
}

func TestUnitDevicePartialParameters(t *testing.T) {
	mock, changeOnGateway := partialDeviceGateway()
	providerFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
			ResourceFactory: NewDeviceResource,
			Client:          mock,
		}),
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			// The gateway's defaults for the keys left out are neither an
			// inconsistent result nor drift (the step's follow-up plan is empty).
			{
				Config: partialDeviceConfig(`{"hostname": "10.0.0.5"}`),
				Check:  resource.TestCheckResourceAttr("ignition_device.test", "parameters", `{"hostname": "10.0.0.5"}`),
			},
			// A nested object is compared by the keys it sets too.
			{
				Config: partialDeviceConfig(`{"hostname":"10.0.0.6","addressing":{"zeroBased":true}}`),
				Check:  resource.TestCheckResourceAttr("ignition_device.test", "parameters", `{"hostname":"10.0.0.6","addressing":{"zeroBased":true}}`),
			},
			// A key the configuration sets, changed on the gateway, is drift.
			{
				PreConfig: func() {
					changeOnGateway(func(c client.DeviceConfig) { c["hostname"] = "10.9.9.9" })
				},
				Config:             partialDeviceConfig(`{"hostname":"10.0.0.6","addressing":{"zeroBased":true}}`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: partialDeviceConfig(`{"hostname":"10.0.0.6","addressing":{"zeroBased":true}}`),
				Check:  resource.TestCheckResourceAttr("ignition_device.test", "parameters", `{"hostname":"10.0.0.6","addressing":{"zeroBased":true}}`),
			},
			// A default changed on the gateway is not: the configuration leaves it to the gateway.
			{
				PreConfig: func() {
					changeOnGateway(func(c client.DeviceConfig) { c["port"] = float64(1502) })
				},
				Config:   partialDeviceConfig(`{"hostname":"10.0.0.6","addressing":{"zeroBased":true}}`),
				PlanOnly: true,
			},
			// Import has no configuration to compare with, so it takes every key.
			{
				ResourceName:  "ignition_device.test",
				ImportState:   true,
				ImportStateId: "Modbus",
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					var got map[string]any
					if err := json.Unmarshal([]byte(states[0].Attributes["parameters"]), &got); err != nil {
						return err
					}
					if got["connectTime"] != float64(10000) || got["port"] != float64(1502) {
						return fmt.Errorf("imported parameters lack the gateway's settings: %v", got)
					}
					return nil
				},
			},
		},
	})
}

func TestUnitProjectParameters(t *testing.T) {
	gateway := map[string]any{
		"a": float64(1),
		"b": map[string]any{"c": true, "d": "x"},
		"e": []any{float64(1), float64(2)},
		"f": "default",
	}
	cases := []struct {
		name, want, projected string
	}{
		{"keeps only the keys set", `{"a":5}`, `{"a":1}`},
		{"recurses into objects", `{"b":{"c":false}}`, `{"b":{"c":true}}`},
		{"takes a list whole", `{"e":[1]}`, `{"e":[1,2]}`},
		{"drops a key the gateway lacks", `{"z":1}`, `{}`},
		{"takes the gateway's value where the type differs", `{"b":"text"}`, `{"b":{"c":true,"d":"x"}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var want any
			if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(projectParameters(gateway, want))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.projected {
				t.Errorf("projectParameters(%s) = %s, want %s", tc.want, got, tc.projected)
			}
		})
	}
}
