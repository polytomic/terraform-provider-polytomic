resource "polytomic_amplitude_connection" "amplitude" {
  name = "example"
  configuration = {
    api_key    = "my-api-key"
    secret_key = "my-secret-key"
  }
}

