---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_user Resource - terraform-provider-polytomic"
subcategory: "Organizations"
description: |-
  A user in a Polytomic organization
---

# polytomic_user (Resource)

A user in a Polytomic organization

## Example Usage

```terraform
resource "polytomic_user" "admin" {
  organization = polytomic_organization.acme.id
  email        = "admin@acmeinc.com"
  role         = "admin"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) Email address
- `organization` (String) Organization ID

### Optional

- `role` (String) Role; one of `user` or `admin`.

### Read-Only

- `id` (String) user identifier


