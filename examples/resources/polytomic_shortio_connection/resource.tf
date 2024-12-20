resource "polytomic_shortio_connection" "shortio" {
  name = "example"
  configuration = {
    api_key = "secret"
  }
}

