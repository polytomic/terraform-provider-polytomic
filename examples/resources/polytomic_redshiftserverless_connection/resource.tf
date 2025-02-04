resource "polytomic_redshiftserverless_connection" "redshiftserverless" {
  name = "example"
  configuration = {
    connection_method = "data_api"
    database          = "users"
    endpoint          = "acme.12345.us-west-2.redshift-serverless.amazonaws.com:5439"
    iam_role_arn      = "arn:aws:iam::012345678910:role/role"
    region            = "us-west-2"
    workgroup         = "default"
  }
}

