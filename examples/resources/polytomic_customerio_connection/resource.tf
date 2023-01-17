resource "polytomic_customerio_connection" "customerio" {
  name = "example"
  configuration = {
    site_id = "my-site-id"
  }
}

