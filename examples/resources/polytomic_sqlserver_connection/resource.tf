resource "polytomic_sqlserver_connection" "sqlserver" {
  organization = polytomic_organization.acme.id
  name         = "Acme, Inc"
  configuration = {
    Hostname = "sqlserver.azure.example.com"
    Username = "polytomic"
    Password = "secret"
    Database = "acme"
    Port     = 1443
  }
}

