resource "polytomic_mongodb_connection" "mongodb" {
  name = "example"
  configuration = {
    hosts    = "mongodb.example.net"
    password = "password"
    username = "admin"
  }
}

