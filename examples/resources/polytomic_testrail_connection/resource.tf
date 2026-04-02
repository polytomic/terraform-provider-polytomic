resource "polytomic_testrail_connection" "testrail" {
  name = "example"
  configuration = {
    hostname = "https://example.testrail.io"
  }
}

