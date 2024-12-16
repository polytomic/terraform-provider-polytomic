resource "polytomic_databricks_connection" "databricks" {
  name = "example"
  configuration = {
    access_token    = "isoz8af6zvp8067gu68gvrp0oftevn"
    cloud_provider  = "aws"
    http_path       = "/sql"
    server_hostname = "dbc-1234dsafas-d0001.cloud.databricks.com"
  }
}

