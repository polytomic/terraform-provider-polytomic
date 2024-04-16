resource "polytomic_klaviyo_connection" "klaviyo" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

