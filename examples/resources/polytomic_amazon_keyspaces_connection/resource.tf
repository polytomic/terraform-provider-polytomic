resource "polytomic_amazon_keyspaces_connection" "amazon_keyspaces" {
  name = "example"
  configuration = {
    access_key_id     = "your-access-key-id"
    auth_method       = "access_key_and_secret"
    region            = "us-east-1"
    secret_access_key = "your-secret-access-key"
    username          = "your-service-username"
  }
}

