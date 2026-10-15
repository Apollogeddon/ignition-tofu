package resources

import (
	"context"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitCoreResources(t *testing.T) {
	mockClient := &client.MockClient{
		CreateProjectFunc: func(ctx context.Context, p client.Project) (*client.Project, error) {
			return &p, nil
		},
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, error) {
			return &client.Project{Name: name, Enabled: base.BoolPtr(true)}, nil
		},
		CreateDatabaseConnectionFunc: func(ctx context.Context, db client.ResourceResponse[client.DatabaseConfig]) (*client.ResourceResponse[client.DatabaseConfig], error) {
			db.Signature = "sig"
			return &db, nil
		},
		GetDatabaseConnectionFunc: func(ctx context.Context, name string) (*client.ResourceResponse[client.DatabaseConfig], error) {
			return &client.ResourceResponse[client.DatabaseConfig]{Name: name, Signature: "sig", Enabled: base.BoolPtr(true), Config: client.DatabaseConfig{Driver: "MySQL", Translator: "mysql", ConnectURL: "jdbc:mysql://localhost"}}, nil
		},
		CreateTagProviderFunc: func(ctx context.Context, tp client.ResourceResponse[client.TagProviderConfig]) (*client.ResourceResponse[client.TagProviderConfig], error) {
			tp.Signature = "sig"
			return &tp, nil
		},
		GetTagProviderFunc: func(ctx context.Context, name string) (*client.ResourceResponse[client.TagProviderConfig], error) {
			return &client.ResourceResponse[client.TagProviderConfig]{Name: name, Signature: "sig", Enabled: base.BoolPtr(true), Config: client.TagProviderConfig{Profile: client.TagProviderProfile{Type: "STANDARD"}}}, nil
		},
		CreateUserSourceFunc: func(ctx context.Context, us client.ResourceResponse[client.UserSourceConfig]) (*client.ResourceResponse[client.UserSourceConfig], error) {
			us.Signature = "sig"
			return &us, nil
		},
		GetUserSourceFunc: func(ctx context.Context, name string) (*client.ResourceResponse[client.UserSourceConfig], error) {
			return &client.ResourceResponse[client.UserSourceConfig]{Name: name, Signature: "sig", Enabled: base.BoolPtr(true), Config: client.UserSourceConfig{Profile: client.UserSourceProfile{Type: "internal"}}}, nil
		},
	}

	providerFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
			Client: mockClient,
			ResourceFactories: []func() fwresource.Resource{
				NewProjectResource,
				NewDatabaseConnectionResource,
				NewTagProviderResource,
				NewUserSourceResource,
			},
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
					resource "ignition_project" "test" {
						name = "project"
					}
					resource "ignition_database_connection" "test" {
						name = "db"
						type = "MySQL"
						connect_url = "jdbc:mysql://localhost"
						translator = "mysql"
					}
					resource "ignition_tag_provider" "test" {
						name = "tags"
						type = "STANDARD"
					}
					resource "ignition_user_source" "test" {
						name = "users"
						type = "internal"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_project.test", "name", "project"),
					resource.TestCheckResourceAttr("ignition_database_connection.test", "type", "MySQL"),
					resource.TestCheckResourceAttr("ignition_tag_provider.test", "name", "tags"),
					resource.TestCheckResourceAttr("ignition_user_source.test", "name", "users"),
				),
			},
		},
	})
}
