resource "polytomic_xero_connection" "xero" {
  name = "example"
  configuration = {
    client_id     = "your-client-id"
    client_secret = "your-client-secret"
  }
}

