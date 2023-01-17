resource "polytomic_redshift_connection" "redshift" {
  name = "example"
  configuration = {
    hostname              = "redshift.example.com"
    username              = "acme"
    password              = "super-secret-password"
    database              = "db"
    port                  = 5439
    aws_access_key_id     = "EXAMPLEKEY"
    aws_secret_access_key = "EXAMPLESECRET"
    s3_bucket_name        = "my-bucket"
    s3_bucket_region      = "us-east-1"
  }
}

