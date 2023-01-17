resource "polytomic_lob_connection" "lob" {
  name = "example"
  configuration = {
    apikey = "my-api-key"
  }
}

