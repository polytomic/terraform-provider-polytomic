---
page_title: "polytomic_netsuitesaconnect_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  NetSuite SuiteAnalytics Connection
---

# polytomic_netsuitesaconnect_connection (Resource)

NetSuite SuiteAnalytics Connection

For detailed configuration guidance, see the [NetSuite SuiteAnalytics connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/netsuitesaconnect).

## Example Usage

```terraform
resource "polytomic_netsuitesaconnect_connection" "netsuitesaconnect" {
  name = "example"
  configuration = {
    account_id = "1234567_SB1"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) NetSuite SuiteAnalytics Connection identifier.
- `force_destroy` (Boolean, Optional) Indicates whether dependent models, syncs, and bulk syncs should be
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

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

- `account_id` (String, Required) Account ID
- `certificate_id` (String, Required) Certificate ID
- `client_id` (String, Required) Client ID
- `private_key` (String, Sensitive, Required) Private key
- `role_id` (String, Required) Role ID

