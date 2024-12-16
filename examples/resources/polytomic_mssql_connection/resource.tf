resource "polytomic_mssql_connection" "mssql" {
  name = "example"
  configuration = {
    database = "sampledb"
    hostname = "example.database.windows.net"
    password = "secret"
    username = "user"
  }
}

