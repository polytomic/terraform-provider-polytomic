resource "polytomic_zendesk_support_connection" "zendesk_support" {
  name = "example"
  configuration = {
    api_token   = "secret-token"
    auth_method = "apitoken"
    domain      = "polytomic.zendesk.com"
    email       = "user@example.com"
  }
}

