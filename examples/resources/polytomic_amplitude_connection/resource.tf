resource "polytomic_amplitude_connection" "amplitude" {
  name = "example"
  configuration = {
    api_key    = "api-key"
    secret_key = "******"
  }
}

