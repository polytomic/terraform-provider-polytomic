resource "polytomic_clickhouse_connection" "clickhouse" {
  name = "example"
  configuration = {
    auth_mode             = "access_key_and_secret"
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    azure_access_key      = "abcdefghijklmnopqrstuvwxyz0123456789/+ABCDEabcdefghijklmnopqrstuvwxyz0123456789/+ABCDE=="
    azure_account_name    = "account"
    cloud_provider        = "aws"
    container_name        = "container"
    database              = "default"
    gcs_bucket_name       = "my-bucket"
    gcs_hmac_access_id    = "GOOG1EXAMPLEACCESSID"
    gcs_hmac_secret       = "bGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9EXAMPLE"
    hostname              = "clickhouse.example.com"
    s3_bucket_name        = "my-bucket"
    s3_bucket_region      = "us-east-1"
    ssh_host              = "bastion.example.com"
    username              = "default"
  }
}

