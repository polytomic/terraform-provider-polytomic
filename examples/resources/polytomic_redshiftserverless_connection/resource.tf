resource "polytomic_redshiftserverless_connection" "redshiftserverless" {
  name = "example"
  configuration = {
    database     = "users"
    iam_role_arn = "arn:aws:iam::012345678910:role/role"
    region       = "us-west-2"
    workgroup    = "default"
  }
}

