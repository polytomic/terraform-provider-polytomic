resource "polytomic_ibm_db2_connection" "ibm_db2" {
  name = "example"
  configuration = {
    account  = "db2admin"
    database = "SAMPLE"
    hostname = "db2.example.com"
    passwd   = "password"
  }
}

