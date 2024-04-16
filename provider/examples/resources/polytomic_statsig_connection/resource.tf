resource "polytomic_statsig_connection" "statsig" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

