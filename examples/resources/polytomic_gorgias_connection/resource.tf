resource "polytomic_gorgias_connection" "gorgias" {
  name = "example"
  configuration = {
    apikey        = "secret-key"
    client_id     = "6218fa8cfe1b2a3c4d5e6f70"
    client_secret = "secret"
    domain        = "acme"
    email         = "user@example.com"
  }
}

