resource "polytomic_ware2go_connection" "ware2go" {
  name = "example"
  configuration = {
    client_id     = "your-client-id"
    client_secret = "your-client-secret"
    merchant_id   = "your-merchant-id"
  }
}

