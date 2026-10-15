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

func TestUnitProjectResource(t *testing.T) {
	mockProject := &client.Project{
		Name:        "test-project",
		Description: "Initial Description",
		Enabled:     base.BoolPtr(true),
	}

	mockClient := &client.MockClient{
		CreateProjectFunc: func(ctx context.Context, p client.Project) (*client.Project, error) {
			mockProject = &p
			return mockProject, nil
		},
		GetProjectFunc: func(ctx context.Context, name string) (*client.Project, error) {
			return mockProject, nil
		},
		UpdateProjectFunc: func(ctx context.Context, p client.Project) (*client.Project, error) {
			mockProject = &p
			return mockProject, nil
		},
		DeleteProjectFunc: func(ctx context.Context, name string) error {
			return nil
		},
	}

	providerFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
			ResourceFactory: NewProjectResource,
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
					resource "ignition_project" "test" {
						name        = "test-project"
						description = "Initial Description"
						enabled     = true
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("ignition_project.test", "description", "Initial Description"),
					resource.TestCheckResourceAttr("ignition_project.test", "enabled", "true"),
				),
			},
			{
				Config: `
					provider "ignition" {
						host  = "http://mock-host"
						token = "mock-token"
					}
					resource "ignition_project" "test" {
						name        = "test-project"
						description = "Updated Description"
						enabled     = false
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_project.test", "description", "Updated Description"),
					resource.TestCheckResourceAttr("ignition_project.test", "enabled", "false"),
				),
			},
		},
	})
}

// enabled and inheritable reach the client as set, false included, and stay unset (left to
// the gateway) while the plan doesn't know them.
func TestProjectMapPlanToClient_Booleans(t *testing.T) {
	r := &ProjectResource{}
	m := ProjectResourceModel{Inheritable: types.BoolValue(false)}
	m.Name, m.Enabled = types.StringValue("p"), types.BoolValue(false)
	p, err := r.MapPlanToClient(context.Background(), &m)
	if err != nil {
		t.Fatal(err)
	}
	if p.Enabled == nil || *p.Enabled || p.Inheritable == nil || *p.Inheritable {
		t.Errorf("Expected enabled and inheritable false, got %v %v", p.Enabled, p.Inheritable)
	}

	m = ProjectResourceModel{Inheritable: types.BoolNull()}
	m.Name, m.Enabled = types.StringValue("p"), types.BoolUnknown()
	if p, _ = r.MapPlanToClient(context.Background(), &m); p.Enabled != nil || p.Inheritable != nil {
		t.Errorf("Expected unknown and null left unset, got %v %v", p.Enabled, p.Inheritable)
	}
}
