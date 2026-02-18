resource "polytomic_stord_connection" "stord" {
  name = "example"
  configuration = {
    api_key         = "******"
    network_id      = "12345"
    organization_id = "67890"
  }
}

