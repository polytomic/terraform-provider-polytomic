package provider

import (
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_organization.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(orgName),
					),
					statecheck.ExpectKnownValue(
						"polytomic_organization.test",
						tfjsonpath.New("sso_domain"),
						knownvalue.StringExact(ssoDomain),
					),
					statecheck.ExpectKnownValue(
						"polytomic_organization.test",
						tfjsonpath.New("sso_org_id"),
						knownvalue.StringExact(ssoOrgId),
					),
					statecheck.ExpectKnownValue(
						"polytomic_organization.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationExists(t, orgName),
				),
			},
			{
				Config: testAccOrganizationResource(orgName2, ssoDomain, ssoOrgId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_organization.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(orgName2),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationExists(t, orgName2),
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
