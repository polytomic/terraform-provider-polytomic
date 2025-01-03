package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccConnectionResource(t *testing.T) {
	name := "TestAccConnection"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionResource(name),
				Check: resource.ComposeTestCheckFunc(
					// Check if the resource exists
					testAccConnectionExists(t, name),
					// Check the name
					resource.TestCheckResourceAttr("polytomic_csv_connection.test", "name", name),
				),
			},
		},
	})
}

func testAccConnectionExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources["polytomic_csv_connection.test"]
		if !ok {
			return fmt.Errorf("not found: polytomic_csv_connection.test")
		}

		client := testClient(t)
		conn, err := client.Connections.Get(context.Background(), resource.Primary.ID)
		if err != nil {
			return err
		}
		assert.Equal(t, name, pointer.Get(conn.Data.Name))

		return nil
	}
}

func testAccConnectionResource(name string) string {
	return fmt.Sprintf(`
resource "polytomic_csv_connection" "test" {
  name         = "%s"
  configuration = {
    url = "https://gist.githubusercontent.com/jpalawaga/20df01c463b82950cc7421e5117a67bc/raw/14bae37fb748114901f7cfdaa5834e4b417537d5/"
  }
}
`, name)
}
