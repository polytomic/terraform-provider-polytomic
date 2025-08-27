resource "polytomic_cloudflare_r2_connection" "cloudflare_r2" {
  name = "example"
  configuration = {
    aws_access_key_id        = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key    = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    bucket_name              = "polytomic/dataset"
    single_table_file_format = "csv"
    single_table_name        = "collection"
  }
}

