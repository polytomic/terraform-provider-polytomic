package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUser_basic(t *testing.T) {
	if APIKey() {
		t.Skip("Skipping test that creates organization resources. To run, use a deployment or partner key.")
	}

	email := "test@example.com"
	email2 := "mIxEdCase@example.com"

	org := "terraform-test-org"
	role := "admin"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResource(email, role, org),
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists(t, email, APIKey()),
				),
			},
			{
				Config: testAccUserResource(email2, role, org),
				Check: resource.ComposeTestCheckFunc(
					testAccUserExists(t, email2, APIKey()),
				),
			},
		},
	})
}

func testAccUserExists(t *testing.T, email string, apiKey bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.acme"]
			if !ok {
				return fmt.Errorf("not found: %s", "polytomic_organization.acme")
			}
			orgID = org.Primary.ID
		}

		rs, ok := s.RootModule().Resources["polytomic_user.admin"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_user.admin")
		}
		if rs.Primary.Attributes["email"] != email {
			return fmt.Errorf("email is %s; want %s", rs.Primary.Attributes["email"], email)
		}

		client := testPartnerClient(t)
		users, err := client.Users.List(t.Context(), orgID)
		if err != nil {
			return err
		}
		var found bool
		for _, user := range users.Data {
			if strings.EqualFold(pointer.Get(user.Email), email) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("user %s not found in API response", email)
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
