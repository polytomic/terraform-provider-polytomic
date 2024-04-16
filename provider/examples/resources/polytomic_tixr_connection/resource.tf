resource "polytomic_tixr_connection" "tixr" {
  name         = "example"
  configuration = {
    client_private_key = "my-client-private-key"
    client_secret = "super-secret"
  }
}

