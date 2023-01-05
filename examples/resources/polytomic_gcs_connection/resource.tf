resource "polytomic_gcs_connection" "gcs" {
  name = "example"
  configuration = {
    project_id = "my-project"
    bucket     = "my-bucket"
  }
}

