resource "polytomic_gcs_connection" "gcs" {
  name = "example"
  configuration = {
    bucket            = "my-bucket"
    single_table_name = "collection"
  }
}

