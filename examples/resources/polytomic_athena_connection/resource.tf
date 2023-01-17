resource "polytomic_athena_connection" "athena" {
  name = "example"
  configuration = {
    access_key_id     = "EXAMPLEACCESSKEYID"
    access_key_secret = "EXAMPLEACCESSKEYSECRET"
    region            = "us-east-1"
    output_bucket     = "athena-output-bucket"
  }
}

