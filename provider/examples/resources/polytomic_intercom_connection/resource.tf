resource "polytomic_intercom_connection" "intercom" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

