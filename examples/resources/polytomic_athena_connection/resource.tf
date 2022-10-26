resource "polytomic_athena_connection" "athena" {
  organization = polytomic_organization.acme.id
  name         = "Acme, Inc"
  configuration = {
    AccessKeyID     = "EXAMPLEACCESSKEYID"
    AccessKeySecret = "EXAMPLEACCESSKEYSECRET"
    Region          = "us-east-1"
    OutputBucket    = "athena-output-bucket"
  }
}

