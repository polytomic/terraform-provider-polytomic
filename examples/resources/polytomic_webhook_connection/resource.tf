resource "polytomic_webhook_connection" "webhook" {
  name = "example"
  configuration = {
    url = "https://example.com/webhook"
  }
}

