resource "polytomic_s3_connection" "s3" {
  organization = polytomic_organization.acme.id
  name         = "example"
  configuration = {
    access_key_id     = "EXAMPLEACCESSKEYID"
    access_key_secret = "EXAMPLEACCESSKEYSECRET"
    region            = "us-east-1"
    bucket            = "my-bucket"
  }
}

