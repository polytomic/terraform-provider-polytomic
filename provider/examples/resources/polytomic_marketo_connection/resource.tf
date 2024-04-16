resource "polytomic_marketo_connection" "marketo" {
  name         = "example"
  configuration = {
    client_id = "my-client-id"
    client_secret = "my-client-secret"
    rest_endpoint = "https://marketo.example.com"
  }
}

