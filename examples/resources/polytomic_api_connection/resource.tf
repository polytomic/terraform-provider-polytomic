resource "polytomic_api_connection" "api" {
  name = "example"
  configuration = {
    body        = jsonencode({ "key" : "value" })
    healthcheck = "https://example.com/healthz"
    url         = "https://example.com"
  }
}

