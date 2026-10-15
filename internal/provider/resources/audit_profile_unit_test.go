package resources

import (
	"context"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitAuditProfileResource(t *testing.T) {
	mockClient := &client.MockClient{
		CreateAuditProfileFunc: func(ctx context.Context, ap client.ResourceResponse[client.AuditProfileConfig]) (*client.ResourceResponse[client.AuditProfileConfig], error) {
			ap.Signature = "sig-123"
			return &ap, nil
		},
		GetAuditProfileFunc: func(ctx context.Context, name string) (*client.ResourceResponse[client.AuditProfileConfig], error) {
			return &client.ResourceResponse[client.AuditProfileConfig]{
				Name:      name,
				Enabled:   base.BoolPtr(true),
				Signature: "sig-123",
				Config: client.AuditProfileConfig{
					Profile: client.AuditProfileProfile{
						Type:          "database",
						RetentionDays: intPtr(90),
					},
					Settings: client.AuditProfileSettings{
						DatabaseName: "IgnitionDB",
						TableName:    "audit_events",
						AutoCreate:   base.BoolPtr(true),
						PruneEnabled: base.BoolPtr(false),
					},
				},
			}, nil
		},
		UpdateAuditProfileFunc: func(ctx context.Context, ap client.ResourceResponse[client.AuditProfileConfig]) (*client.ResourceResponse[client.AuditProfileConfig], error) {
			ap.Signature = "sig-456"
			return &ap, nil
		},
		DeleteAuditProfileFunc: func(ctx context.Context, name, signature string) error {
			return nil
		},
	}

	providerFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
			ResourceFactory: NewAuditProfileResource,
			Client:          mockClient,
		}),
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "ignition" {
						host  = "http://mock-host"
						token = "mock-token"
					}
					resource "ignition_audit_profile" "test" {
						name        = "TestAuditProfile"
						type        = "database"
						database    = "IgnitionDB"
						table_name  = "audit_events"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_audit_profile.test", "name", "TestAuditProfile"),
					resource.TestCheckResourceAttr("ignition_audit_profile.test", "type", "database"),
					resource.TestCheckResourceAttr("ignition_audit_profile.test", "database", "IgnitionDB"),
					resource.TestCheckResourceAttr("ignition_audit_profile.test", "table_name", "audit_events"),
				),
			},
		},
	})
}

func intPtr(n int) *int { return &n }

// retention_days = 0 (no pruning) and false booleans reach the client; unknown ones don't.
func TestAuditProfileMapPlanToClient_ZeroAndFalse(t *testing.T) {
	r := &AuditProfileResource{}
	m := AuditProfileResourceModel{
		Type: types.StringValue("database"), RetentionDays: types.Int64Value(0),
		PruneEnabled: types.BoolValue(false), AutoCreate: types.BoolValue(false), EnableStoreAndForward: types.BoolUnknown(),
	}
	c, err := r.MapPlanToClient(context.Background(), &m)
	if err != nil {
		t.Fatal(err)
	}
	if c.Profile.RetentionDays == nil || *c.Profile.RetentionDays != 0 {
		t.Errorf("Expected retention 0, got %v", c.Profile.RetentionDays)
	}
	if c.Settings.PruneEnabled == nil || *c.Settings.PruneEnabled || c.Settings.AutoCreate == nil || *c.Settings.AutoCreate {
		t.Errorf("Expected prune and auto-create false, got %v %v", c.Settings.PruneEnabled, c.Settings.AutoCreate)
	}
	if c.Settings.EnableStoreAndForward != nil {
		t.Errorf("Expected an unknown value left unset, got %v", *c.Settings.EnableStoreAndForward)
	}
}
