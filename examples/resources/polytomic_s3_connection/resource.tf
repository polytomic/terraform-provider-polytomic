resource "polytomic_s3_connection" "s3" {
  name = "example"
  configuration = {
    access_key_id     = "EXAMPLEACCESSKEYID"
    access_key_secret = "EXAMPLEACCESSKEYSECRET"
    region            = "us-east-1"
    bucket            = "my-bucket"
  }
}

