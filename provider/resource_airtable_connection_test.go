// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAirtableConnection(t *testing.T) {
	name := "TestAcc"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Missing configuration
			{
				Config:      testAccAirtableConnectionResourceNoConfig(name),
				ExpectError: regexp.MustCompile("The argument \"configuration\" is required, but no definition was found."),
			},
			// Empty configuration
			{
				Config:      testAccAirtableConnectionResourceMissingRequired(name),
				ExpectError: regexp.MustCompile("Inappropriate value for attribute \"configuration\":"),
			},
			{
				Config: testAccAirtableConnectionResource(name),
				Check: resource.ComposeTestCheckFunc(
					// Check if the resource exists
					testAccAirtableConnectionExists(name),
					// Check the name
					resource.TestCheckResourceAttr("polytomic_airtable_connection.test", "name", name),
				),
			},
		},
	})
}

func testAccAirtableConnectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources["polytomic_airtable_connection.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_airtable_connection.test")
		}

		client := testClient()
		connections, err := client.Connections().List(context.TODO())
		if err != nil {
			return err
		}
		var found bool
		for _, connection := range connections {
			if connection.Name == name {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("connection %s not found", name)
		}

		return nil

	}
}

func testAccAirtableConnectionResource(name string) string {
	return fmt.Sprintf(`
resource "polytomic_airtable_connection" "test" {
	name          = "%s"
	configuration = {
   api_key = "my-api-key"
   }
}
`, name)
}

func testAccAirtableConnectionResourceNoConfig(name string) string {
	return fmt.Sprintf(`
resource "polytomic_airtable_connection" "test" {
	name          = "%s"
}
`, name)
}

func testAccAirtableConnectionResourceMissingRequired(name string) string {
	return fmt.Sprintf(`
resource "polytomic_airtable_connection" "test" {
	name          = "%s"
	 configuration = {}
}`, name)
}
