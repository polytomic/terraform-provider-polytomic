resource "polytomic_databricks_connection" "databricks" {
  name = "example"
  configuration = {
    server_hostname       = "https://my.databricks.com"
    port                  = 443
    access_token          = "my-access-token"
    http_path             = "/sql"
    aws_access_key_id     = "EXAMPLEKEY"
    aws_secret_access_key = "EXAMPLESECRET"
    s3_bucket_name        = "my-bucket"
    s3_bucket_region      = "us-east-1"
  }
}

