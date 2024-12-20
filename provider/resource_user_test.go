package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUser_basic(t *testing.T) {
	if os.Getenv("TEST_ORG_RESOURCES") != "true" {
		t.Skip("Skipping test that creates resources in the Terraform test organization. To run, set TEST_ORG_RESOURCES=true")
	}

	email := "test@example.com"
	email2 := "mIxEdCase@example.com"

	org := "terraform-test-org"
	role := "admin"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResource(email, role, org),
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists(email),
				),
			},
			{
				Config: testAccUserResource(email2, role, org),
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists(email2),
				),
			},
		},
	})
}

func testAccUserExists(email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["polytomic_user.admin"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_user.admin")
		}
		if rs.Primary.Attributes["email"] != email {
			return fmt.Errorf("email is %s; want %s", rs.Primary.Attributes["email"], email)
		}
		return nil

	}
}

func testAccUserResource(email, role, organization string) string {
	return fmt.Sprintf(`
resource "polytomic_user" "admin" {
	organization = polytomic_organization.acme.id
	email        = "%s"
	role         = "%s"
}
resource "polytomic_organization" "acme" {
	name       = "%s"
	sso_domain = "acmeinc.com"
}
`, email, role, organization)
}
