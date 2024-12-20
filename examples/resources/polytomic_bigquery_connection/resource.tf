resource "polytomic_bigquery_connection" "bigquery" {
  name = "example"
  configuration = {
    location = "us-east1"
  }
}

