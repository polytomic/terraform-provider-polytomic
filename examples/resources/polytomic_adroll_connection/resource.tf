resource "polytomic_adroll_connection" "adroll" {
  name = "example"
  configuration = {
    client_id             = "your_client_id"
    client_secret         = "your_client_secret"
    personal_access_token = "secret"
  }
}

