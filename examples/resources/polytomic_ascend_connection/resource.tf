resource "polytomic_ascend_connection" "ascend" {
  name = "example"
  configuration = {
    api_key = "secret"
  }
}

