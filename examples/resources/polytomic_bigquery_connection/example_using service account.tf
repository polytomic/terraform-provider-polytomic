
resource "polytomic_bigquery_connection" "bigquery" {
  name = "BigQuery (Service Account)"

  configuration = {
    auth_method     = "service_account_key"
    project_id      = "my-gcp-project"
    service_account = file(var.bq_service_account_json_key_path)
  }
}
