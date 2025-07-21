package roundtrip

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/polytomic/terraform-provider-polytomic/provider"
)

// Test connection round-trip
func TestAccRoundTrip_Connection(t *testing.T) {
	connName := provider.ValidName(provider.ToSnakeCase(fmt.Sprintf("test-%s", uuid.NewString())))
	var testOrgName string
	if !provider.APIKey() {
		testOrgName = connName
	}
	resourceName := fmt.Sprintf("polytomic_csv_connection.%s", connName)
	tfConfig := provider.TestCaseTfResource(t, connectionResourceTemplate, provider.TestCaseTfArgs{
		Name:   connName,
		APIKey: provider.APIKey(),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 PreCheck(t),
		ProtoV6ProviderFactories: provider.GetTestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Step 1: Create the connection
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", connName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					ImportAndValidate(
						t.Context(),
						[]string{resourceName},
						RoundTripOptions{
							ValidateSensitive: false,
							IgnoreFields: []string{
								"created_at",
								"updated_at",
							},
							OrgName: testOrgName,
						},
					)),
			},
			// Step 2: Test refreshing state via reading
			{
				RefreshState: true,
			},
			// Step 3: Test importing, if we're using an API key
			{
				SkipFunc: func() (bool, error) {
					return !provider.APIKey(), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configuration.url", // May change during test
				},
				ResourceName: resourceName,
			},
		},
	})
}

const connectionResourceTemplate = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
  name = "{{.Name}}"
}
{{end}}
resource "polytomic_csv_connection" "{{.Name}}" {
  name          = "{{.Name}}"
  {{if not .APIKey}}
  organization  = polytomic_organization.test.id
  {{end}}
  configuration = {
    url = "https://gist.githubusercontent.com/jpalawaga/20df01c463b82950cc7421e5117a67bc/raw/14bae37fb748114901f7cfdaa5834e4b417537d5/"
  }
}
`
