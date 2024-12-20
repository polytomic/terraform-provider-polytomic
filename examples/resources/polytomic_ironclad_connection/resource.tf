resource "polytomic_ironclad_connection" "ironclad" {
  name = "example"
  configuration = {
    api_key       = "secret"
    client_id     = "ironclad"
    client_secret = "secret"
    user_as_email = "email@domain.com"
  }
}

