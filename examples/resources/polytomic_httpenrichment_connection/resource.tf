resource "polytomic_httpenrichment_connection" "httpenrichment" {
  name = "example"
  configuration = {
    body        = jsonencode({ "key" : "value" })
    healthcheck = "https://example.com/healthz"
    url         = "https://example.com"
  }
}

