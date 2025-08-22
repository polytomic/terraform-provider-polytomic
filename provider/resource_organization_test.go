package provider

import (
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOrganization_basic(t *testing.T) {
	if APIKey() {
		t.Skip("Skipping test that creates organization resources. To run, use a deployment or partner key.")
	}

	orgName := "terraform-test-org"
	orgName2 := "terraform-test-org-updated"
	ssoDomain := "acmeinc.com"
	ssoOrgId := "org_123456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationResource(orgName, ssoDomain, ssoOrgId),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationExists(t, orgName),
					resource.TestCheckResourceAttr("polytomic_organization.test", "name", orgName),
					resource.TestCheckResourceAttr("polytomic_organization.test", "sso_domain", ssoDomain),
					resource.TestCheckResourceAttr("polytomic_organization.test", "sso_org_id", ssoOrgId),
				),
			},
			{
				Config: testAccOrganizationResource(orgName2, ssoDomain, ssoOrgId),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationExists(t, orgName2),
					resource.TestCheckResourceAttr("polytomic_organization.test", "name", orgName2),
				),
			},
		},
	})
}

func testAccOrganizationExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["polytomic_organization.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_organization.test")
		}
		if rs.Primary.Attributes["name"] != name {
			return fmt.Errorf("name is %s; want %s", rs.Primary.Attributes["name"], name)
		}

		client := testPartnerClient(t)
		organization, err := client.Organization.Get(t.Context(), rs.Primary.ID)
		if err != nil {
			return err
		}
		if pointer.Get(organization.Data.Name) != name {
			return fmt.Errorf("organization name is %s; want %s", pointer.Get(organization.Data.Name), name)
		}
		return nil
	}
}

func testAccOrganizationResource(name, ssoDomain, ssoOrgId string) string {
	return fmt.Sprintf(`
resource "polytomic_organization" "test" {
	name       = "%s"
	sso_domain = "%s"
	sso_org_id = "%s"
}
`, name, ssoDomain, ssoOrgId)
}
