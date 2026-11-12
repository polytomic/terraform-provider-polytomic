resource "polytomic_aws_cloudwatch_metrics_connection" "aws_cloudwatch_metrics" {
  name = "example"
  configuration = {
    access_id         = "AKIAIOSFODNN7EXAMPLE"
    auth_mode         = "access_key_and_secret"
    region            = "us-east-1"
    secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
}

