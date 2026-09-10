resource "polytomic_amazon_rds_postgresql_connection" "amazon_rds_postgresql" {
  name = "example"
  configuration = {
    database     = "sampledb"
    hostname     = "database.123456789012.us-east-1.rds.amazonaws.com"
    iam_role_arn = "arn:aws:iam::123456789012:role/polytomic-rds"
    publication  = "polytomic"
    region       = "us-east-1"
    ssh_host     = "bastion.example.com"
    username     = "polytomic"
  }
}

