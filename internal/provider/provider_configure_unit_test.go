package provider

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// configure runs Configure on the real provider with the given attribute values; nil is null.
func configure(t *testing.T, deferralAllowed bool, host, token, timeout tftypes.Value) *provider.ConfigureResponse {
	t.Helper()
	ctx := context.Background()
	p := New("test")()
	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)
	raw := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		"host":               host,
		"token":              token,
		"allow_insecure_tls": tftypes.NewValue(tftypes.Bool, nil),
		"request_timeout":    timeout,
	})
	resp := &provider.ConfigureResponse{}
	p.Configure(ctx, provider.ConfigureRequest{
		Config:             tfsdk.Config{Schema: schemaResp.Schema, Raw: raw},
		ClientCapabilities: provider.ConfigureProviderClientCapabilities{DeferralAllowed: deferralAllowed},
	}, resp)
	return resp
}

var (
	unknownString = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	nullNumber    = tftypes.NewValue(tftypes.Number, nil)
)

func TestUnitConfigureUnknownHostDefers(t *testing.T) {
	resp := configure(t, true, unknownString, tftypes.NewValue(tftypes.String, "t"), nullNumber)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Expected no error, got %v", resp.Diagnostics)
	}
	if resp.Deferred == nil || resp.Deferred.Reason != provider.DeferredReasonProviderConfigUnknown {
		t.Errorf("Expected a deferral for unknown configuration, got %+v", resp.Deferred)
	}
}

func TestUnitConfigureUnknownTokenWithoutDeferral(t *testing.T) {
	resp := configure(t, false, tftypes.NewValue(tftypes.String, "http://gw:8088"), unknownString, nullNumber)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Expected no error, got %v", resp.Diagnostics)
	}
	if resp.Diagnostics.WarningsCount() != 1 || resp.ResourceData != nil || resp.DataSourceData != nil {
		t.Errorf("Expected one warning and no client, got %v / %v", resp.Diagnostics, resp.ResourceData)
	}
}

func TestUnitConfigureRequestTimeout(t *testing.T) {
	resp := configure(t, false, tftypes.NewValue(tftypes.String, "http://gw:8088"), tftypes.NewValue(tftypes.String, "t"), tftypes.NewValue(tftypes.Number, 5))
	if resp.Diagnostics.HasError() {
		t.Fatalf("Expected no error, got %v", resp.Diagnostics)
	}
	c, ok := resp.ResourceData.(*client.Client)
	if !ok || c.RequestTimeout != 5*time.Second {
		t.Errorf("Expected a client with a 5s request timeout, got %+v", resp.ResourceData)
	}

	resp = configure(t, false, tftypes.NewValue(tftypes.String, "http://gw:8088"), tftypes.NewValue(tftypes.String, "t"), nullNumber)
	if c, ok := resp.ResourceData.(*client.Client); !ok || c.RequestTimeout != client.DefaultRequestTimeout {
		t.Errorf("Expected the default request timeout, got %+v", resp.ResourceData)
	}
}

func TestUnitRequestTimeoutRange(t *testing.T) {
	for _, v := range []string{"0", "-1", "3601"} {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{{
				Config: `
provider "ignition" {
  host            = "http://gw:8088"
  token           = "t"
  request_timeout = ` + v + `
}
data "ignition_project" "p" {
  name = "p"
}
`,
				ExpectError: regexp.MustCompile(`request_timeout`),
			}},
		})
	}
}

// A host another resource sets plans and applies: the provider is configured at apply,
// and a later plan with the host unknown again keeps the existing resource's state.
func TestUnitHostKnownOnlyAtApply(t *testing.T) {
	var gets int
	mock := &client.MockClient{
		CreateUserSourceFunc: func(_ context.Context, us client.ResourceResponse[client.UserSourceConfig]) (*client.ResourceResponse[client.UserSourceConfig], error) {
			us.Signature = "sig"
			return &us, nil
		},
		GetUserSourceFunc: func(_ context.Context, name string) (*client.ResourceResponse[client.UserSourceConfig], error) {
			gets++
			return &client.ResourceResponse[client.UserSourceConfig]{
				Name: name, Enabled: base.BoolPtr(true), Signature: "sig",
				Config: client.UserSourceConfig{Profile: client.UserSourceProfile{Type: "INTERNAL"}},
			}, nil
		},
		DeleteUserSourceFunc: func(context.Context, string, string) error { return nil },
	}
	factories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&IgnitionProvider{version: "test", client: mock}),
	}
	config := func(host string) string {
		return `
resource "terraform_data" "host" {
  input = "` + host + `"
}
provider "ignition" {
  host  = terraform_data.host.output
  token = "t"
}
resource "ignition_user_source" "test" {
  name = "us"
  type = "INTERNAL"
}
`
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: config("http://gw-1:8088"),
				Check:  resource.TestCheckResourceAttr("ignition_user_source.test", "signature", "sig"),
			},
			{
				Config: config("http://gw-2:8088"),
				Check:  resource.TestCheckResourceAttr("ignition_user_source.test", "type", "INTERNAL"),
			},
		},
	})
	if gets == 0 {
		t.Error("Expected the user source to be read once the host was known")
	}
}
