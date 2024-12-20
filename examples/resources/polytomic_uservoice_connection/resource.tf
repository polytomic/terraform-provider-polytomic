resource "polytomic_uservoice_connection" "uservoice" {
  name = "example"
  configuration = {
    api_key = "secret"
    domain  = "polytomic"
  }
}

