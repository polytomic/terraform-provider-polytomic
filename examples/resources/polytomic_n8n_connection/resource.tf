resource "polytomic_n8n_connection" "n8n" {
  name = "example"
  configuration = {
    url = "https://your-instance.app.n8n.cloud"
  }
}

