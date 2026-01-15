resource "polytomic_awsathena_connection" "awsathena" {
  name = "example"
  configuration = {
    access_id         = "AKIAIOSFODNN7EXAMPLE"
    auth_mode         = "access_key_and_secret"
    outputbucket      = "s3://polytomic-athena-results/customer-dataset"
    region            = "us-east-1"
    secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
}

