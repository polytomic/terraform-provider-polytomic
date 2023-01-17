resource "polytomic_s3_connection" "s3" {
  name = "example"
  configuration = {
    aws_access_key_id     = "EXAMPLEACCESSKEYID"
    aws_secret_access_key = "EXAMPLEACCESSKEYSECRET"
    s3_bucket_region      = "us-east-1"
    s3_bucket_name        = "my-bucket"
  }
}

