resource "polytomic_cloudsql_connection" "cloudsql" {
  name = "example"
  configuration = {
    connection_name = "my-project:us-central1:my-instance"
    credentials     = "data.account_credentials.json"
    database        = "my-db"
    username        = "cloudsql"
  }
}

