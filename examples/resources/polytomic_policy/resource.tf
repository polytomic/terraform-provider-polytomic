resource "polytomic_policy" "example" {
  name = "Terraform role"
  policy_actions = [
    {
      action = "create"
      role_ids = [
        polytomic_role.example.id
      ]
    },
    {
      action = "delete"
      role_ids = [
        polytomic_role.example.id
      ]
    },
  ]
}
