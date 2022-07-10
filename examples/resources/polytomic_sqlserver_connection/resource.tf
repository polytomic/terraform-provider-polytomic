resource "polytomic_sqlserver_connection" "sqlserver" {
  workspace = polytomic_workspace.acme.id
  name      = "Acme, Inc"
  configuration = {
    hostname = "sqlserver.azure.example.com"
    username = "polytomic"
    password = ""
    database = "acme"
    port     = 1443
  }
}
