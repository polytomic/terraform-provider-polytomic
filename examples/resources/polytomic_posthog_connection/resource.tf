resource "polytomic_posthog_connection" "posthog" {
  name = "example"
  configuration = {
    api_key  = "secret"
    location = "us"
  }
}

