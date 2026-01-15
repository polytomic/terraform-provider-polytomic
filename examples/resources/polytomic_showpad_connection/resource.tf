resource "polytomic_showpad_connection" "showpad" {
  name = "example"
  configuration = {
    subdomain = "acme"
  }
}

