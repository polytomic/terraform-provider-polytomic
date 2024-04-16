resource "polytomic_postgresql_connection" "postgresql" {
  name         = "example"
  configuration = {
    hostname = "acme.postgres.database.example.com"
    username = "acme"
    password = "1234567890"
    database = "acme"
    port = 5432
  }
}

