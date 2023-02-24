resource "polytomic_uservoice_connection" "uservoice" {
  name = "example"
  configuration = {
    api_key = "my-api-key"
    domain  = "example.com"
  }
}

