resource "polytomic_walmart_marketplace_connection" "walmart_marketplace" {
  name = "example"
  configuration = {
    client_id     = "your-client-id"
    client_secret = "your-client-secret"
  }
}

