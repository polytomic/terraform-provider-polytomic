resource "polytomic_redshiftserverless_connection" "redshiftserverless" {
  name = "example"
  configuration = {
    database     = "dev"
    workgroup    = "default-workgroup"
    iam_role_arn = "arn:aws:iam::XXXX:role/polytomic-redshiftserverless"
    external_id  = "db"
  }
}

