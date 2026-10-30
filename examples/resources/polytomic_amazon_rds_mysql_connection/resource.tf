resource "polytomic_amazon_rds_mysql_connection" "amazon_rds_mysql" {
  name = "example"
  configuration = {
    database              = "sampledb"
    hostname              = "database.123456789012.us-east-1.rds.amazonaws.com"
    iam_role_arn          = "arn:aws:iam::123456789012:role/polytomic-rds"
    rds_endpoint_hostname = "database.123456789012.us-east-1.rds.amazonaws.com"
    region                = "us-east-1"
    ssh_host              = "bastion.example.com"
    username              = "polytomic"
  }
}

