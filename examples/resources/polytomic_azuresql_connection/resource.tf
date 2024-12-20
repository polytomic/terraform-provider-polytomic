resource "polytomic_azuresql_connection" "azuresql" {
  name = "example"
  configuration = {
    database = "sampledb"
    hostname = "example.database.windows.net"
    password = "secret"
    username = "user"
  }
}

