package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestUnitUserSourceResource(t *testing.T) {
	mockClient := &client.MockClient{
		CreateUserSourceFunc: func(ctx context.Context, us client.ResourceResponse[client.UserSourceConfig]) (*client.ResourceResponse[client.UserSourceConfig], error) {
			us.Signature = "sig-123"
			return &us, nil
		},
		GetUserSourceFunc: func(ctx context.Context, name string) (*client.ResourceResponse[client.UserSourceConfig], error) {
			return &client.ResourceResponse[client.UserSourceConfig]{
				Name:      name,
				Enabled:   base.BoolPtr(true),
				Signature: "sig-123",
				Config: client.UserSourceConfig{
					Profile: client.UserSourceProfile{
						Type:         "INTERNAL",
						FailoverMode: "Soft",
					},
				},
			}, nil
		},
		UpdateUserSourceFunc: func(ctx context.Context, us client.ResourceResponse[client.UserSourceConfig]) (*client.ResourceResponse[client.UserSourceConfig], error) {
			us.Signature = "sig-456"
			return &us, nil
		},
		DeleteUserSourceFunc: func(ctx context.Context, name, signature string) error {
			return nil
		},
	}

	providerFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
			ResourceFactory: NewUserSourceResource,
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
					resource "ignition_user_source" "test" {
						name          = "TestUserSource"
						type          = "INTERNAL"
						failover_mode = "Soft"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_user_source.test", "name", "TestUserSource"),
					resource.TestCheckResourceAttr("ignition_user_source.test", "type", "INTERNAL"),
					resource.TestCheckResourceAttr("ignition_user_source.test", "failover_mode", "Soft"),
				),
			},
		},
	})
}

// userSourceGateway is a mock gateway holding one user source, recording what each create
// and update sent.
type userSourceGateway struct {
	stored client.ResourceResponse[client.UserSourceConfig]
	sent   []client.ResourceResponse[client.UserSourceConfig]
}

func (g *userSourceGateway) client() *client.MockClient {
	save := func(us client.ResourceResponse[client.UserSourceConfig]) (*client.ResourceResponse[client.UserSourceConfig], error) {
		g.sent = append(g.sent, us)
		// a field left out of the request keeps the gateway's value
		if b, err := json.Marshal(us); err != nil || !strings.Contains(string(b), `"description":`) {
			us.Description = g.stored.Description
		}
		if us.Config.Profile.ScheduleRestricted == nil {
			us.Config.Profile.ScheduleRestricted = g.stored.Config.Profile.ScheduleRestricted
		}
		us.Signature = "sig"
		g.stored = us
		return &us, nil
	}
	return &client.MockClient{
		CreateUserSourceFunc: func(_ context.Context, us client.ResourceResponse[client.UserSourceConfig]) (*client.ResourceResponse[client.UserSourceConfig], error) {
			return save(us)
		},
		UpdateUserSourceFunc: func(_ context.Context, us client.ResourceResponse[client.UserSourceConfig]) (*client.ResourceResponse[client.UserSourceConfig], error) {
			return save(us)
		},
		GetUserSourceFunc: func(_ context.Context, name string) (*client.ResourceResponse[client.UserSourceConfig], error) {
			us := g.stored
			return &us, nil
		},
		DeleteUserSourceFunc: func(context.Context, string, string) error { return nil },
	}
}

func userSourceConfig(attrs string) string {
	return `
provider "ignition" {}
resource "ignition_user_source" "test" {
  name = "us"
  type = "INTERNAL"
` + attrs + `
}
`
}

// Left out, schedule_restricted is not sent, and the gateway's value shows without a diff;
// a configured false is sent.
func TestUnitUserSourceScheduleRestrictedFollowsGateway(t *testing.T) {
	tr := true
	g := &userSourceGateway{}
	g.stored.Config.Profile.ScheduleRestricted = &tr
	factories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{ResourceFactory: NewUserSourceResource, Client: g.client()}),
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: userSourceConfig(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_user_source.test", "schedule_restricted", "true"),
					func(*terraform.State) error {
						if g.sent[0].Config.Profile.ScheduleRestricted != nil {
							return fmt.Errorf("create sent schedule_restricted %v", *g.sent[0].Config.Profile.ScheduleRestricted)
						}
						return nil
					},
				),
			},
			{Config: userSourceConfig(""), PlanOnly: true},
			{
				Config: userSourceConfig("schedule_restricted = false"),
				Check:  resource.TestCheckResourceAttr("ignition_user_source.test", "schedule_restricted", "false"),
			},
			// a configuration from before the default was dropped, now without the field: no diff
			{Config: userSourceConfig(""), PlanOnly: true},
		},
	})
	last := g.sent[len(g.sent)-1].Config.Profile.ScheduleRestricted
	if last == nil || *last {
		t.Errorf("Expected the update to send false, got %v", last)
	}
}

// A description removed from the configuration is sent cleared, so state matches the plan.
func TestUnitUserSourceDescriptionRemoved(t *testing.T) {
	g := &userSourceGateway{}
	factories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{ResourceFactory: NewUserSourceResource, Client: g.client()}),
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: userSourceConfig(`description = "a description"`),
				Check:  resource.TestCheckResourceAttr("ignition_user_source.test", "description", "a description"),
			},
			{
				Config: userSourceConfig(""),
				Check:  resource.TestCheckNoResourceAttr("ignition_user_source.test", "description"),
			},
		},
	})
}
