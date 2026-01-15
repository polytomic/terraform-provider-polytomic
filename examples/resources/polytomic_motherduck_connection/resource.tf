resource "polytomic_motherduck_connection" "motherduck" {
  name = "example"
  configuration = {
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    database              = "my_db"
    s3_bucket_name        = "s3://polytomic/dataset"
    s3_bucket_region      = "us-east-1"
  }
}

