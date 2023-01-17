resource "polytomic_mongodb_connection" "mongodb" {
  name = "example"
  configuration = {
    hosts    = "mongodb.example.com"
    username = "user"
    password = "secret"
    database = "db"
  }
}

