resource "polytomic_datadog_connection" "datadog" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

