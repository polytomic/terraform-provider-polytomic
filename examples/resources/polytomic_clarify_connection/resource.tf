resource "polytomic_clarify_connection" "clarify" {
  name = "example"
  configuration = {
    server_url = "https://api.dev.clarify.ai"
    workspace  = "acme"
  }
}

