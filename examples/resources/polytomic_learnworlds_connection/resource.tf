resource "polytomic_learnworlds_connection" "learnworlds" {
  name = "example"
  configuration = {
    school_url = "https://my-school.learnworlds.com"
  }
}

