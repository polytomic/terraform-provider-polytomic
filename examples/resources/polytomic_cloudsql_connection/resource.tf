resource "polytomic_cloudsql_connection" "cloudsql" {
  name = "example"
  configuration = {
    connection_name = "my-project:us-central1:my-instance"
    database        = "my-db"
    username        = "cloudsql"
  }
}

