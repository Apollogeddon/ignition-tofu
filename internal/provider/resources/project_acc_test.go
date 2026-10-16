package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	if os.Getenv("IGNITION_HOST") == "" || os.Getenv("IGNITION_TOKEN") == "" {
		t.Skip("Skipping acceptance test: IGNITION_HOST and/or IGNITION_TOKEN not set")
	}

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_project.test", "name", rName),
				),
			},
			{
				ResourceName:      "ignition_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// an update that turns booleans off and sets a description must reach the gateway
			{
				Config: testAccProjectResourceConfigUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_project.test", "enabled", "false"),
					resource.TestCheckResourceAttr("ignition_project.test", "inheritable", "false"),
					resource.TestCheckResourceAttr("ignition_project.test", "description", "updated by the acceptance test"),
					resource.TestCheckResourceAttr("ignition_project.test", "title", "Acceptance"),
				),
			},
			{
				ResourceName:      "ignition_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// fields removed from the configuration are cleared on the gateway
			{
				Config: testAccProjectResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("ignition_project.test", "description"),
					resource.TestCheckNoResourceAttr("ignition_project.test", "title"),
				),
			},
		},
	})
}

func testAccProjectResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "ignition" {}

resource "ignition_project" "test" {
  name        = %[1]q
}
`, name)
}

func testAccProjectResourceConfigUpdated(name string) string {
	return fmt.Sprintf(`
provider "ignition" {}

resource "ignition_project" "test" {
  name        = %[1]q
  description = "updated by the acceptance test"
  title       = "Acceptance"
  enabled     = false
  inheritable = false
}
`, name)
}
