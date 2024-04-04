resource "polytomic_lob_connection" "lob" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

