resource "polytomic_googlecloudmysql_connection" "googlecloudmysql" {
  name = "example"
  configuration = {
    connection_name = "project:region:instance"
    database        = "sampledb"
    password        = "secret"
    username        = "cloudsql"
  }
}

