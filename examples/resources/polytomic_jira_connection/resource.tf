resource "polytomic_jira_connection" "jira" {
  name = "example"
  configuration = {
    access_token = "secret"
    auth_method  = "pat"
    url          = "https://jira.mycompany.com/"
  }
}

