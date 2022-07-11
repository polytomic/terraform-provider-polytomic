resource "polytomic_athena_connection" "acmelake" {
  organization = polytomic_organization.acme.id
  name         = "Acme, Inc Data"
  configuration = {
    access_key_id     = ""
    access_key_secret = ""
    region            = "us-west-2"
    output_bucket     = "acme-data-athena-output"
  }
}
