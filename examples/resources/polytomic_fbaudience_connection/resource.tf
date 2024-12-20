resource "polytomic_fbaudience_connection" "fbaudience" {
  name = "example"
  configuration = {
    account_id        = "1234567890"
    auth_method       = "token"
    byo_app_token     = "secret"
    graph_api_version = "v19.0"
  }
}

