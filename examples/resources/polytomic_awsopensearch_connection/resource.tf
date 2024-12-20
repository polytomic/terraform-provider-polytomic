resource "polytomic_awsopensearch_connection" "awsopensearch" {
  name = "example"
  configuration = {
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    endpoint              = "es.us-east-2.amazonaws.com"
    region                = "us-east-1"
  }
}

