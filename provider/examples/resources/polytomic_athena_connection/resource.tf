resource "polytomic_athena_connection" "athena" {
  name         = "example"
  configuration = {
    access_id = "EXAMPLEACCESSKEYID"
    secret_access_key = "EXAMPLEACCESSKEYSECRET"
    region = "us-east-1"
    outputbucket = "athena-output-bucket"
  }
}

