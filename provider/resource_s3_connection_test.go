package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccS3(t *testing.T) {
	key := "test"
	secret := "test"
	region := "us-east-1"
	bucket := "test-bucket"
	org := "terraform-test-org"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccS3Resource(key, secret, region, bucket, org),
				Check: resource.ComposeTestCheckFunc(
					testAccS3Exists(bucket),
				),
			},
		},
	})
}

func testAccS3Exists(bucket string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["polytomic_s3_connection.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_s3_connection.test")
		}
		if rs.Primary.Attributes["configuration.bucket"] != bucket {
			return fmt.Errorf("bucket is %s; want %s", rs.Primary.Attributes["configuration.bucket"], bucket)
		}
		return nil

	}
}

func testAccS3Resource(accessKey, secretKey, region, bucket, organization string) string {
	return fmt.Sprintf(`
resource "polytomic_s3_connection" "test" {
	organization = polytomic_organization.acme.id
	name         = "Acc Test Bucket"
	configuration = {
	  access_key_id     = "%s"
	  access_key_secret = "%s"
	  region            = "%s"
	  bucket            = "%s"
	}
}
resource "polytomic_organization" "acme" {
	name       = "%s"
	sso_domain = "acmeinc.com"
}
`, accessKey, secretKey, region, bucket, organization)
}
