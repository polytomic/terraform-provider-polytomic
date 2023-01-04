resource "polytomic_sqlserver_connection" "sqlserver" {
  name = "example"
  configuration = {
    hostname = "sqlserver.azure.example.com"
    username = "polytomic"
    database = "acme"
    port     = 1443
  }
}

