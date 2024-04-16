resource "polytomic_dialpad_connection" "dialpad" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

