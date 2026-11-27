---
page_title: "polytomic_sap_business_one_for_hana_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  SAP Business One for HANA Connection
---

# polytomic_sap_business_one_for_hana_connection (Resource)

SAP Business One for HANA Connection

For detailed configuration guidance, see the [SAP Business One for HANA connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/sap_business_one_for_hana).

## Example Usage

```terraform
resource "polytomic_sap_business_one_for_hana_connection" "sap_business_one_for_hana" {
  name = "example"
  configuration = {
    hostname = "hana.example.com"
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

- `id` (String) SAP Business One for HANA Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `hostname` (String) Hostname or IP address of the SAP HANA server.
- `password` (String, Sensitive) Password of the SAP HANA database user.
- `port` (Number) SQL port for the SAP HANA database.
- `username` (String) Username of the SAP HANA database user.

#### Optional

- `disable_tls` (Boolean) Disable TLS

    Whether to connect without TLS. By default, the connection uses TLS and verifies the server certificate.


