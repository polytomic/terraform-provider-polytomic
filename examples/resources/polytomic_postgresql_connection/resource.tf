resource "polytomic_postgresql_connection" "postgresql" {
  name = "example"
  configuration = {
    hostname = "acme.postgres.database.example.com"
    username = "acme"
    database = "acme"
    port     = 5432
  }
}

