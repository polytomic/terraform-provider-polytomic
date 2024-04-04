resource "polytomic_customerio_connection" "customerio" {
  name         = "example"
  configuration = {
    site_id = "my-site-id"
    tracking_api_key = "my-tracking-api-key"
    app_api_key = "my-app-api-key"
  }
}

