resource "polytomic_azureblob_connection" "azureblob" {
  organization = polytomic_organization.acme.id
  name         = "example"
  configuration = {
    account_name   = "my-account"
    access_key     = "abcdefghijklmnopqrstuvwxyz0123456789=="
    container_name = "my-container"
  }
}

