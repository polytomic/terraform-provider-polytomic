resource "polytomic_zoominfo_connection" "zoominfo" {
  name = "example"
  configuration = {
    username    = "my-username"
    client_id   = "my-client-id"
    private_key = "my-private-key"
  }
}

