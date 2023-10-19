resource "polytomic_mailercheck_connection" "mailercheck" {
  name = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

