package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceResource(t *testing.T) {
	if os.Getenv("IGNITION_HOST") == "" || os.Getenv("IGNITION_TOKEN") == "" {
		t.Skip("Skipping acceptance test: IGNITION_HOST and/or IGNITION_TOKEN not set")
	}

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing, with every setting the driver has.
			{
				Config: testAccDeviceResourceConfig(rName, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_device.test", "name", rName),
					resource.TestCheckResourceAttr("ignition_device.test", "type", "ProgrammableSimulatorDevice"),
					resource.TestCheckResourceAttr("ignition_device.test", "parameters", `{"legacyMode":false,"repeat":false,"timeIntervalRate":1000}`),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ignition_device.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDeviceResourceConfig(rName, 2000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_device.test", "parameters", `{"legacyMode":false,"repeat":false,"timeIntervalRate":2000}`),
				),
			},
			{
				ResourceName:      "ignition_device.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Partial parameters: the gateway's defaults for the settings left
			// out are neither an inconsistent result nor drift.
			{
				Config: testAccDevicePartialConfig(rName, 3000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ignition_device.test", "parameters", `{"timeIntervalRate":3000}`),
				),
			},
		},
	})
}

func testAccDevicePartialConfig(name string, timeIntervalRate int) string {
	return fmt.Sprintf(`
provider "ignition" {}

resource "ignition_device" "test" {
  name       = %[1]q
  type       = "ProgrammableSimulatorDevice"
  parameters = "{\"timeIntervalRate\":%[2]d}"
}
`, name, timeIntervalRate)
}

func testAccDeviceResourceConfig(name string, timeIntervalRate int) string {
	return fmt.Sprintf(`
provider "ignition" {}

resource "ignition_device" "test" {
  name       = %[1]q
  type       = "ProgrammableSimulatorDevice"
  parameters = "{\"legacyMode\":false,\"repeat\":false,\"timeIntervalRate\":%[2]d}"
}
`, name, timeIntervalRate)
}
