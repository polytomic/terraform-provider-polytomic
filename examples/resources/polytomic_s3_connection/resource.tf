resource "polytomic_s3_connection" "s3" {
  name = "example"
  configuration = {
    aws_access_key_id        = "EXAMPLEACCESSKEYID"
    aws_secret_access_key    = "EXAMPLEACCESSKEYSECRET"
    s3_bucket_region         = "us-east-1"
    s3_bucket_name           = "my-bucket"
    is_single_table          = true
    is_directory_snapshot    = true
    dir_glob_pattern         = "data/tables/*/*.csv"
    single_table_name        = "my_table"
    single_table_file_format = "csv"
    skip_lines               = 1
  }
}

