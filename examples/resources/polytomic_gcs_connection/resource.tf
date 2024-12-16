resource "polytomic_gcs_connection" "gcs" {
  name = "example"
  configuration = {
    bucket = "my-bucket"
  }
}

