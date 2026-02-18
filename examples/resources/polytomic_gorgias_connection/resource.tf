resource "polytomic_gorgias_connection" "gorgias" {
  name = "example"
  configuration = {
    apikey = "secret-key"
    domain = "acme"
    email  = "user@example.com"
  }
}

