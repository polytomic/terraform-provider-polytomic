resource "polytomic_gcs_connection" "gcs" {
  name = "example"
  configuration = {
    project_id                  = "my-project"
    service_account_credentials = data.account_credentials.json
    bucket                      = "my-bucket"
  }
}

