---
page_title: "polytomic_googleworkspace_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Workspace Connection
---

# polytomic_googleworkspace_connection (Resource)

Google Workspace Connection

For detailed configuration guidance, see the [Google Workspace connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/googleworkspace).

## Example Usage

```terraform
resource "polytomic_googleworkspace_connection" "googleworkspace" {
  name = "example"
  configuration = {
    client_id     = "eb669428-1854-4cb1-a560-403e05b8acbf"
    client_secret = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
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

- `id` (String) Google Workspace Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `auth_method` (String) Authentication method

    Default: browser Valid values: <code>oauth</code> (OAuth), <code>service_account</code> (Service Account). Default: <code>oauth</code>.

#### Optional

- `client_id` (String, Sensitive)
- `client_secret` (String, Sensitive)
- `customer_id` (String) Customer ID
- `service_account` (String, Sensitive) Service account key

#### Read-Only

- `client_email` (String) Connected Account


