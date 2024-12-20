resource "polytomic_sftp_connection" "sftp" {
  name = "example"
  configuration = {
    path     = "/path/to/files"
    ssh_host = "sftp.example.net"
    ssh_user = "user"
  }
}

