---
page_title: "polytomic_polytomic_metadata_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Polytomic Metadata Connection
---

# polytomic_polytomic_metadata_connection (Resource)

Polytomic Metadata Connection

For detailed configuration guidance, see the [Polytomic Metadata connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/polytomic_metadata).

## Example Usage

```terraform
resource "polytomic_polytomic_metadata_connection" "polytomic_metadata" {
  name = "example"
  configuration = {
    auth_mode = "personal_api_key"
  }
}
```

## Schema

### Required

- `name` (String)
- `configuration` (Attributes) See [below for nested schema](#nestedatt--configuration).

### Optional

- `organization` (String) Organization ID.
- `force_destroy` (Boolean) Indicates whether dependent models, syncs, and bulk syncs should be
cascade-deleted when this connection is destroyed.

    This only deletes other resources when the connection is destroyed, not when
setting this parameter to `true`. Once this parameter is set to `true`, there
must be a successful `terraform apply` run before a destroy is required to
update this value in the resource state. Without a successful `terraform apply`
after this parameter is set, this flag will have no effect. If setting this
field in the same operation that would require replacing the connection or
destroying the connection, this flag will not work. Additionally when importing
a connection, a successful `terraform apply` is required to set this value in
state before it will take effect on a destroy operation.

### Read-Only

- `id` (String) Polytomic Metadata Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `auth_mode` (String) Authentication Method

    Type of API key to use for authentication Valid values: <code>personal_api_key</code> (Personal API Key), <code>partner_api_key</code> (Partner API Key). Default: <code>personal_api_key</code>.

#### Optional

- `deployment_api_key` (String, Sensitive) Deployment API Key
- `partner_api_key` (String, Sensitive) Partner API Key

    Partner API key provided by Polytomic
- `personal_api_key` (String, Sensitive) Personal API Key

    Your personal API key from Polytomic settings

#### Read-Only

- `connected_org` (String) Connected organization
- `connected_user` (String) Connected user


