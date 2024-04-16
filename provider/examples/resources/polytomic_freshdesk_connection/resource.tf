resource "polytomic_freshdesk_connection" "freshdesk" {
  name         = "example"
  configuration = {
    apikey = "my-api-key"
    subdomain = "example.freshdesk.com"
  }
}

