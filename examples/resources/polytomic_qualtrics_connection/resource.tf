resource "polytomic_qualtrics_connection" "qualtrics" {
  name = "example"
  configuration = {
    api_key = "secret"
  }
}

