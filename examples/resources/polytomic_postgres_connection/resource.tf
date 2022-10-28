resource "polytomic_postgres_connection" "postgres" {
  organization = polytomic_organization.acme.id
  name         = "example"
  configuration = {
    hostname = "acme.postgres.database.example.com"
    username = "acme"
    password = "1234567890"
    database = "acme"
    port     = 5432
  }
}

