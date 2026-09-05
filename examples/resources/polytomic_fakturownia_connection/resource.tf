resource "polytomic_fakturownia_connection" "fakturownia" {
  name = "example"
  configuration = {
    subdomain = "acme"
  }
}

