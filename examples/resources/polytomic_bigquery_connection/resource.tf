resource "polytomic_bigquery_connection" "bigquery" {
  name = "example"
  configuration = {
    auth_method = "service_account_key"
    bucket      = "my-bucket"
    location    = "us-east1"
  }
}

