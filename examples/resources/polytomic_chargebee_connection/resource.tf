resource "polytomic_chargebee_connection" "chargebee" {
  name = "example"
  configuration = {
    api_key         = "secret"
    product_catalog = "2.0"
    site            = "https://example.chargebee.com"
  }
}

