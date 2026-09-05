resource "polytomic_meridian_connection" "meridian" {
  name = "example"
  configuration = {
    client_id     = "your_client_id"
    client_secret = "your_client_secret"
  }
}

