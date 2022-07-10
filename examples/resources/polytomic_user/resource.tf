resource "polytomic_user" "admin" {
  workspace = polytomic_workspace.acme.id
  email     = "admin@acmeinc.com"
  role      = "admin"
}
