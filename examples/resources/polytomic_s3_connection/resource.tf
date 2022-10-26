resource "polytomic_s3_connection" "s3" {
  organization = polytomic_organization.acme.id
  name         = "Acme, Inc"
  configuration = {
    AccessKeyID     = "EXAMPLEACCESSKEYID"
    AccessKeySecret = "EXAMPLEACCESSKEYSECRET"
    Region          = "us-east-1"
    Bucket          = "my-bucket"
  }
}

