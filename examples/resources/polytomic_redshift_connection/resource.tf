resource "polytomic_redshift_connection" "redshift" {
  name = "example"
  configuration = {
    hostname          = "redshift.example.com"
    username          = "acme"
    database          = "db"
    port              = 5439
    aws_access_key_id = "EXAMPLEKEY"
    s3_bucket_name    = "my-bucket"
    s3_bucket_region  = "us-east-1"
  }
}

