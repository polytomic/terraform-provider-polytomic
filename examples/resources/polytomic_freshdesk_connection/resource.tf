resource "polytomic_freshdesk_connection" "freshdesk" {
  name = "example"
  configuration = {
    subdomain = "example.freshdesk.com"
  }
}

