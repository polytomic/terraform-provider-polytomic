resource "polytomic_athena_connection" "athena" {
  name = "example"
  configuration = {
    access_id         = "EXAMPLEACCESSKEYID"
    outputbucket      = "athena-output-bucket"
    region            = "us-east-1"
    secret_access_key = "EXAMPLEACCESSKEYSECRET"
  }
}

