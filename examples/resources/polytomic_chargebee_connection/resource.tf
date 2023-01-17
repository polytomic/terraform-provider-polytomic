resource "polytomic_chargebee_connection" "chargebee" {
  name = "example"
  configuration = {
    site = "site.example.com"
  }
}

