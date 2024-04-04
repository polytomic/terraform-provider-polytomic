resource "polytomic_gcs_connection" "gcs" {
  name         = "example"
  configuration = {
    project_id = "my-project"
    service_account = "data.account_credentials.json"
    bucket = "my-bucket"
  }
}

