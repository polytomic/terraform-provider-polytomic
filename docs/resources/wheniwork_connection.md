---
page_title: "polytomic_wheniwork_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  When I Work Connection
---

# polytomic_wheniwork_connection (Resource)

When I Work Connection

For detailed configuration guidance, see the [When I Work connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/wheniwork).

## Example Usage

```terraform
resource "polytomic_wheniwork_connection" "wheniwork" {
  name = "example"
  configuration = {
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

- `id` (String) When I Work Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `developer_key` (String, Sensitive) Developer key

    The When I Work developer key used to authenticate login requests.
- `email` (String) The email address of the When I Work person to authenticate.
- `password` (String, Sensitive) The password of the When I Work person to authenticate.

#### Optional

- `end_date` (String) Coverage end date

    The end boundary for date-bounded collections; defaults to 90 days after the read, with at most 20 years of coverage.
- `start_date` (String) History start date

    The inclusive start date for date-bounded collections; defaults to 365 days before the read.
- `workplace_user_id` (String) Workplace user ID

    The optional workplace user ID sent as W-UserId to select the account.


