resource "polytomic_replyio_connection" "replyio" {
  name = "example"
  configuration = {
    api_key = "secret"
  }
}

