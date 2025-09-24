resource "polytomic_s3_connection" "s3" {
  name = "example"
  configuration = {
    auth_mode                = "iam_role"
    aws_access_key_id        = "EXAMPLEACCESSKEYID"
    aws_secret_access_key    = "EXAMPLEACCESSKEYSECRET"
    iam_role_arn             = "arn:aws:iam::XXXX:role/polytomic-s3"
    external_id              = "00000000-0000-0000-0000-000000000000"
    s3_bucket_region         = "us-east-1"
    s3_bucket_name           = "my-bucket"
    is_single_table          = true
    is_directory_snapshot    = true
    directory_glob_pattern   = "data/tables/*/*.csv"
    single_table_name        = "my_table"
    single_table_file_format = "csv"
    skip_lines               = 1
  }
}

