resource "polytomic_cloudflare_logs_connection" "cloudflare_logs" {
  name = "example"
  configuration = {
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    bucket_name           = "polytomic/dataset"
  }
}

