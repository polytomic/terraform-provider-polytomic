resource "polytomic_redshiftserverless_connection" "redshiftserverless" {
  name = "example"
  configuration = {
    database            = "dev"
    workgroup           = "default-workgroup"
    region              = "us-east-1"
    iam_role_arn        = "arn:aws:iam::XXXX:role/polytomic-redshiftserverless"
    external_id         = "db"
    connection_method   = "endpoint"
    serverless_endpoint = "acme.12345.us-west-2.redshift-serverless.amazonaws.com:5439"
    override_endpoint   = true
    data_api_endpoint   = "https://redshift-data.us-west-2.amazonaws.com"
    use_unload          = true
    s3_bucket_name      = "my-bucket"
    s3_bucket_region    = "us-east-1"
  }
}

