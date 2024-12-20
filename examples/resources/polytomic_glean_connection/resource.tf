resource "polytomic_glean_connection" "glean" {
  name = "example"
  configuration = {
    api_key = "secret"
    domain  = "customer"
  }
}

