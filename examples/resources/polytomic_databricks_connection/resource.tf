resource "polytomic_databricks_connection" "databricks" {
  name = "example"
  configuration = {
    access_token             = "isoz8af6zvp8067gu68gvrp0oftevn"
    auth_mode                = "access_key_and_secret"
    aws_access_key_id        = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key    = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    azure_access_key         = "abcdefghijklmnopqrstuvwxyz0123456789/+ABCDEabcdefghijklmnopqrstuvwxyz0123456789/+ABCDE=="
    azure_account_name       = "account"
    cloud_provider           = "aws"
    container_name           = "container"
    databricks_auth_mode     = "access_token"
    http_path                = "/sql"
    s3_bucket_name           = "s3://polytomic-databricks-results/customer-dataset"
    s3_bucket_region         = "us-east-1"
    server_hostname          = "dbc-1234dsafas-d0001.cloud.databricks.com"
    service_principal_id     = "sp-1234abcd"
    service_principal_secret = "abcdefghijklmnopqrstuvwxyz0123456789/+ABCDEabcdefghijklmnopqrstuvwxyz0123456789/+ABCDE=="
    ssh_host                 = "bastion.example.com"
  }
}

