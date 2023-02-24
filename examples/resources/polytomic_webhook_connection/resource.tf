resource "polytomic_webhook_connection" "example" {
  name = "Example"
  configuration = {
    url = "https://example.com/webhook"
  }
}
