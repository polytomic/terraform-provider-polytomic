resource "polytomic_m3ter_connection" "m3ter" {
  name = "example"
  configuration = {
    client_id     = "your-access-key-id"
    client_secret = "your-api-secret"
    org_id        = "your-org-id-uuid"
  }
}

