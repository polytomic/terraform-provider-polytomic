resource "polytomic_sftp_connection" "sftp" {
  name = "example"
  configuration = {
    auth_mode = "private_key"
    path      = "/path/to/files"
    ssh_host  = "sftp.example.net"
    ssh_user  = "user"
  }
}

