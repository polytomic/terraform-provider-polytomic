resource "polytomic_gcs_connection" "gcs" {
  organization = polytomic_organization.acme.id
  name         = "example"
  configuration = {
    project_id                  = "my-project"
    service_account_credentials = data.account_credentials.json
    bucket                      = "my-bucket"
  }
}

