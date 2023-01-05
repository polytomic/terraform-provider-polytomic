resource "polytomic_s3_connection" "s3" {
  name = "example"
  configuration = {
    region = "us-east-1"
    bucket = "my-bucket"
  }
}

