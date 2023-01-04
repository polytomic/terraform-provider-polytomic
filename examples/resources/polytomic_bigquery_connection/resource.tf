resource "polytomic_bigquery_connection" "bigquery" {
  name = "example"
  configuration = {
    project_id = "my-project"
    location   = "us-central1"
  }
}

