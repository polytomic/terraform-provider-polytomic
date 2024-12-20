resource "polytomic_constructionwire_connection" "constructionwire" {
  name = "example"
  configuration = {
    email    = "user@example.com"
    password = "password"
  }
}

