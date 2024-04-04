resource "polytomic_survicate_connection" "survicate" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

