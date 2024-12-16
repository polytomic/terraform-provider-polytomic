resource "polytomic_shopify_connection" "shopify" {
  name = "example"
  configuration = {
    admin_api_token = "secret"
    store           = "store"
  }
}

