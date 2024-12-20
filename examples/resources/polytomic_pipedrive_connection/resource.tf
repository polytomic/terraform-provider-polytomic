resource "polytomic_pipedrive_connection" "pipedrive" {
  name = "example"
  configuration = {
    api_key = "secret"
    domain  = "polytomic"
  }
}

