resource "polytomic_pinterest_ads_connection" "pinterest_ads" {
  name = "example"
  configuration = {
    client_id     = "your_client_id"
    client_secret = "your_client_secret"
  }
}

