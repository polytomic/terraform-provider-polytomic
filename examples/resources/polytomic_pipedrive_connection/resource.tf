resource "polytomic_pipedrive_connection" "pipedrive" {
  name = "example"
  configuration = {
    domain = "my-domain.example.com"
  }
}

