resource "polytomic_slack_connection" "slack" {
  name = "example"
  configuration = {
    api_key = "secret"
  }
}

