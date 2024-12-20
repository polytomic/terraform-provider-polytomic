resource "polytomic_redshift_connection" "redshift" {
  name = "example"
  configuration = {
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    database              = "mydb"
    hostname              = "mycluster.us-west-2.redshift.amazonaws.com"
    password              = "password"
    s3_bucket_name        = "my-bucket"
    s3_bucket_region      = "us-west-2"
    ssh_host              = "bastion.example.com"
    username              = "redshift_user"
  }
}

