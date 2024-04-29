resource "polytomic_redshiftserverless_connection" "redshiftserverless" {
  name = "example"
  configuration = {
    database          = "dev"
    workgroup         = "default-workgroup"
    iam_role_arn      = "arn:aws:iam::XXXX:role/polytomic-redshiftserverless"
    external_id       = "db"
    override_endpoint = true
    data_api_endpoint = "https://redshift-data.us-west-2.amazonaws.com"
  }
}

