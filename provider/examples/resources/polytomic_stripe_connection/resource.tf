resource "polytomic_stripe_connection" "stripe" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

