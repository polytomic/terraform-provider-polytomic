resource "polytomic_s3_connection" "s3" {
  name = "example"
  configuration = {
    auth_mode        = "access_key_and_secret"
    s3_bucket_name   = "s3://polytomic/dataset"
    s3_bucket_region = "us-east-1"
  }
}

