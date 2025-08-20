resource "polytomic_freshservice_connection" "freshservice" {
  name = "example"
  configuration = {
    subdomain = "polytomic"
  }
}

