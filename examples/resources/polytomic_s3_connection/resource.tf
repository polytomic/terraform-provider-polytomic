resource "polytomic_s3_connection" "s3" {
  name = "example"
  configuration = {
    s3_bucket_region = "us-east-1"
    s3_bucket_name   = "my-bucket"
  }
}

