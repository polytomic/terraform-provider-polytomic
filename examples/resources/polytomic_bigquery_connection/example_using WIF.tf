
resource "polytomic_bigquery_connection" "bigquery" {
  name = "BigQuery (WIF)"

  configuration = {
    auth_method       = "workload_identity_federation"
    wif_project_id    = "my-gcp-project"
    credential_config = file(var.bq_credential_config_path)
  }
}
