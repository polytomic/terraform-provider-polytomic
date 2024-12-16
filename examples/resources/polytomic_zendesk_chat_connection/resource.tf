resource "polytomic_zendesk_chat_connection" "zendesk_chat" {
  name = "example"
  configuration = {
    domain = "polytomic.zendesk.com"
  }
}

