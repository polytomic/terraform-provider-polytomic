resource "polytomic_googlecloudsql_connection" "googlecloudsql" {
  name = "example"
  configuration = {
    connection_name = "project:region:instance"
    database        = "sampledb"
    password        = "secret"
    publication     = "polytomic"
    username        = "cloudsql"
  }
}

