package provider

import (
	"context"
	"regexp"
	"strings"
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

// Without deferral (OpenTofu) an unknown attribute is an error naming it: planning without the
// gateway would leave existing resources unrefreshed.
func TestUnitConfigureUnknownTokenWithoutDeferral(t *testing.T) {
	resp := configure(t, false, tftypes.NewValue(tftypes.String, "http://gw:8088"), unknownString, nullNumber)
	if resp.Diagnostics.ErrorsCount() != 1 || resp.ResourceData != nil || resp.DataSourceData != nil || resp.Deferred != nil {
		t.Fatalf("Expected one error and no client, got %v / %v", resp.Diagnostics, resp.ResourceData)
	}
	if d := resp.Diagnostics.Errors()[0]; !strings.Contains(d.Detail(), "token") || !strings.Contains(d.Detail(), "-target") {
		t.Errorf("Expected the error to name token and the way out, got %q", d.Detail())
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

// On OpenTofu, which can't defer, a host another resource sets is refused rather than
// planned against an unrefreshed state: a gateway replaced in the same run would otherwise
// be left without the resources state says it has.
func TestUnitHostKnownOnlyAtApply(t *testing.T) {
	var creates int
	mock := &client.MockClient{
		CreateUserSourceFunc: func(_ context.Context, us client.ResourceResponse[client.UserSourceConfig]) (*client.ResourceResponse[client.UserSourceConfig], error) {
			creates++
			us.Signature = "sig"
			return &us, nil
		},
		GetUserSourceFunc: func(_ context.Context, name string) (*client.ResourceResponse[client.UserSourceConfig], error) {
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
	known := `
provider "ignition" {
  host  = "http://gw-1:8088"
  token = "t"
}
`
	userSource := `
resource "ignition_user_source" "test" {
  name = "us"
  type = "INTERNAL"
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: known + userSource,
				Check:  resource.TestCheckResourceAttr("ignition_user_source.test", "signature", "sig"),
			},
			// the gateway is replaced in the same run: its new address isn't known at plan time
			{
				Config: `
resource "terraform_data" "host" {
  input = "http://gw-2:8088"
}
provider "ignition" {
  host  = terraform_data.host.output
  token = "t"
}
` + userSource,
				ExpectError: regexp.MustCompile(`(?s)host comes from a value that is only known after apply.*-target`),
			},
			// back on a known host, so the test's destroy can reach the gateway
			{Config: known + userSource, PlanOnly: true},
		},
	})
	if creates != 1 {
		t.Errorf("Expected only the first step to create, got %d creates", creates)
	}
}
