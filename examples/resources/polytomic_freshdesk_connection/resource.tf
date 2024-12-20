resource "polytomic_freshdesk_connection" "freshdesk" {
  name = "example"
  configuration = {
    apikey    = "secret"
    subdomain = "polytomic"
  }
}

