resource "polytomic_fbaudience_connection" "fbaudience" {
  name = "example"
  configuration = {
    auth_method   = "token"
    byo_app_token = "secret"
  }
}

