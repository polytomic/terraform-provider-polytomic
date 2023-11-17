resource "polytomic_awsopensearch_connection" "awsopensearch" {
  name = "example"
  configuration = {
    endpoint              = "https://example-domain-123abcdefg.us-west-2.es.amazonaws.com"
    aws_access_key_id     = "EXAMPLEACCESSKEYID"
    aws_secret_access_key = "EXAMPLEACCESSKEYSECRET"
    region                = "us-east-1"
  }
}

