resource "polytomic_marketo_connection" "marketo" {
  name = "example"
  configuration = {
    client_id     = "629b6d74-f602-47f4-8fef-388485343d85"
    client_secret = "123*******************xyz"
    rest_endpoint = "https://123-ABC-999.mktorest.com/rest"
  }
}

