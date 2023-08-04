resource "polytomic_dynamodb_connection" "dynamodb" {
  name = "example"
  configuration = {
    access_id         = "my-access-key-id"
    secret_access_key = "my-secret-access-key"
    region            = "us-east-1"
  }
}

