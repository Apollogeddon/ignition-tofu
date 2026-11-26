package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func historianMock(t *testing.T, stored *client.ResourceResponse[client.HistorianProviderConfig]) *client.MockClient {
	t.Helper()
	return &client.MockClient{
		CreateHistorianProviderFunc: func(ctx context.Context, item client.ResourceResponse[client.HistorianProviderConfig]) (*client.ResourceResponse[client.HistorianProviderConfig], error) {
			if item.Module != "com.inductiveautomation.historian" || item.Type != "historian-provider" {
				return nil, fmt.Errorf("expected the historian module's historian-provider, got %s/%s", item.Module, item.Type)
			}
			item.Signature = "sig-1"
			*stored = item
			return &item, nil
		},
		GetHistorianProviderFunc: func(ctx context.Context, name string) (*client.ResourceResponse[client.HistorianProviderConfig], error) {
			if name != stored.Name {
				return nil, fmt.Errorf("not found")
			}
			found := *stored
			return &found, nil
		},
		DeleteHistorianProviderFunc: func(ctx context.Context, name, signature string) error {
			return nil
		},
	}
}

func TestUnitHistorianProviderResource_CreateWithDefaults(t *testing.T) {
	var stored client.ResourceResponse[client.HistorianProviderConfig]
	providerFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
			ResourceFactory: NewHistorianProviderResource,
			Client:          historianMock(t, &stored),
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
					resource "ignition_historian_provider" "unit" {
						name     = "demo"
						database = "demo"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_historian_provider.unit", "type", "SqlHistorian"),
					resource.TestCheckResourceAttr("ignition_historian_provider.unit", "partition_units", "MONTH"),
					resource.TestCheckResourceAttr("ignition_historian_provider.unit", "pruning_enabled", "false"),
					resource.TestCheckResourceAttr("ignition_historian_provider.unit", "signature", "sig-1"),
				),
			},
		},
	})

	settings := stored.Config.Settings
	if stored.Config.Profile.Type != "SqlHistorian" || settings.Database != "demo" {
		t.Fatalf("sent %+v", stored.Config)
	}
	// every setting is sent, with the gateway's own defaults for the hidden ones
	if !settings.Partition.Enabled || settings.Partition.Size != 1 || settings.Partition.PartitionSeedQueryLimit != 2 ||
		settings.Partition.OptimizedWindowSeconds != 60 || !settings.TrackSce || settings.StaleMultiplier != 2 {
		t.Fatalf("defaults not sent: %+v", settings)
	}
}

func TestUnitHistorianProviderResource_Pruning(t *testing.T) {
	var stored client.ResourceResponse[client.HistorianProviderConfig]
	providerFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
			ResourceFactory: NewHistorianProviderResource,
			Client:          historianMock(t, &stored),
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
					resource "ignition_historian_provider" "unit" {
						name            = "plant"
						database        = "history_db"
						partition_size  = 1
						partition_units = "WEEK"
						pruning_enabled = true
						pruning_age     = 90
						pruning_units   = "DAY"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_historian_provider.unit", "pruning_age", "90"),
					resource.TestCheckResourceAttr("ignition_historian_provider.unit", "partition_units", "WEEK"),
				),
			},
		},
	})

	if p := stored.Config.Settings.Pruning; !p.Enabled || p.Age != 90 || p.AgeUnits != "DAY" {
		t.Fatalf("pruning sent as %+v", p)
	}
}
