resource "polytomic_pipedrive_connection" "pipedrive" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
    domain = "my-domain.example.com"
  }
}

