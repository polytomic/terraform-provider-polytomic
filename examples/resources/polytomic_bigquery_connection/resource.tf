resource "polytomic_bigquery_connection" "bigquery" {
  name = "example"
  configuration = {
    bucket   = "my-bucket"
    location = "us-east1"
  }
}

