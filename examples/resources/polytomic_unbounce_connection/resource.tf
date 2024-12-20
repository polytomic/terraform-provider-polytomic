resource "polytomic_unbounce_connection" "unbounce" {
  name = "example"
  configuration = {
    api_key = "secret"
  }
}

