package provider

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGlobalErrorSubscribersResource(t *testing.T) {
	name := fmt.Sprintf("TestAccGlobalErrorSubscribers-%s", uuid.NewString())
	email1 := fmt.Sprintf("%s-1@example.com", name)
	email2 := fmt.Sprintf("%s-2@example.com", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, globalErrorSubscribersResourceTemplateTwoEmails, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_global_error_subscribers.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(globalErrorSubscribersResourceID),
					),
					statecheck.ExpectKnownValue(
						"polytomic_global_error_subscribers.test",
						tfjsonpath.New("emails"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(email1),
							knownvalue.StringExact(email2),
						}),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccGlobalErrorSubscribersMatch(t, []string{email1, email2}, APIKey()),
				),
			},
			{
				Config: TestCaseTfResource(t, globalErrorSubscribersResourceTemplateOneEmail, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_global_error_subscribers.test",
						tfjsonpath.New("emails"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(email1),
						}),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccGlobalErrorSubscribersMatch(t, []string{email1}, APIKey()),
				),
			},
		},
	})
}

func testAccGlobalErrorSubscribersMatch(t *testing.T, expected []string, apiKey bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: %s", "polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}

		_, ok := s.RootModule().Resources["polytomic_global_error_subscribers.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_global_error_subscribers.test")
		}

		client := testClient(t, orgID)
		resp, err := client.Notifications.GetGlobalErrorSubscribers(t.Context())
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("empty response")
		}

		actual := resp.Emails
		if actual == nil {
			actual = []string{}
		}

		if !sameStringSet(expected, actual) {
			return fmt.Errorf("expected emails %v, got %v", expected, actual)
		}

		return nil
	}
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := map[string]int{}
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		m[s]--
		if m[s] < 0 {
			return false
		}
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

const globalErrorSubscribersResourceTemplateTwoEmails = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
  name = "{{.Name}}"
}
{{end}}

resource "polytomic_global_error_subscribers" "test" {
  emails = [
    "{{.Name}}-1@example.com",
    "{{.Name}}-2@example.com",
  ]
{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}
`

const globalErrorSubscribersResourceTemplateOneEmail = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
  name = "{{.Name}}"
}
{{end}}

resource "polytomic_global_error_subscribers" "test" {
  emails = [
    "{{.Name}}-1@example.com",
  ]
{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}
`
