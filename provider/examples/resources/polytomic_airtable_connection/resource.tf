resource "polytomic_airtable_connection" "airtable" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

