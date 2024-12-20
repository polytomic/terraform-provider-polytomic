resource "polytomic_predictleads_connection" "predictleads" {
  name = "example"
  configuration = {
    api_key   = "token"
    api_token = "key"
  }
}

