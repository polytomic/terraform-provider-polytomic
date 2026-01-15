resource "polytomic_tiktok_ads_connection" "tiktok_ads" {
  name = "example"
  configuration = {
    client_id     = "your_client_id"
    client_secret = "your_client_secret"
  }
}

