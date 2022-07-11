resource "polytomic_user" "admin" {
  organization = polytomic_organization.acme.id
  email        = "admin@acmeinc.com"
  role         = "admin"
}
