resource "polytomic_tigris_connection" "tigris" {
  name = "example"
  configuration = {
    aws_access_key_id        = "tid_..."
    aws_secret_access_key    = "tsec_..."
    bucket_name              = "polytomic/dataset"
    single_table_file_format = "csv"
    single_table_name        = "collection"
  }
}

