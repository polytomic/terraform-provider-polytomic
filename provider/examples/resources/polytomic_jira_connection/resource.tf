resource "polytomic_jira_connection" "jira" {
  name         = "example"
  configuration = {
    url = "https://example.atlassian.net"
    auth_method = "apikey/pat"
    username = "user"
    api_key = "my-api-key"
    access_token = "my-access-token"
  }
}

