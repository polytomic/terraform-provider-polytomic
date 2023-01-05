resource "polytomic_athena_connection" "athena" {
  name = "example"
  configuration = {
    region        = "us-east-1"
    output_bucket = "athena-output-bucket"
  }
}

