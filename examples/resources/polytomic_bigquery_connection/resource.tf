resource "polytomic_bigquery_connection" "bigquery" {
  name = "example"
  configuration = {
    project_id      = "my-project"
    service_account = "data.account_credentials.json"
    location        = "us-central1"
  }
}

