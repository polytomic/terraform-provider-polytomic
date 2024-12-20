resource "polytomic_sqlserver_connection" "sqlserver" {
  name = "example"
  configuration = {
    database = "acme"
    hostname = "sqlserver.azure.example.com"
    password = "secret"
    port     = 1443
    ssl      = true
    username = "polytomic"
  }
}

