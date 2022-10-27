resource "polytomic_sqlserver_connection" "sqlserver" {
  organization = polytomic_organization.acme.id
  name         = "example"
  configuration = {
    hostname = "sqlserver.azure.example.com"
    username = "polytomic"
    password = "secret"
    database = "acme"
    port     = 1443
  }
}

