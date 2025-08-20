resource "polytomic_mssql_connection" "mssql" {
  name = "example"
  configuration = {
    database = "sampledb"
    hostname = "example.database.windows.net"
    password = "secret"
    ssh_host = "bastion.example.com"
    username = "user"
  }
}

