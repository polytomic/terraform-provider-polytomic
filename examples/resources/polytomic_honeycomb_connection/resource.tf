resource "polytomic_honeycomb_connection" "honeycomb" {
  name = "example"
  configuration = {
    api_key = "secret"
    dataset = "dataset"
  }
}

