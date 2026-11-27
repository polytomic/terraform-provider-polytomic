resource "polytomic_customeriowarehouseexports_connection" "customeriowarehouseexports" {
  name = "example"
  configuration = {
    auth_mode             = "access_key_and_secret"
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    cloud_provider        = "aws"
    gcs_bucket_name       = "gs://my-bucket/exports"
    s3_bucket_name        = "s3://polytomic/dataset"
    s3_bucket_region      = "us-east-1"
  }
}

