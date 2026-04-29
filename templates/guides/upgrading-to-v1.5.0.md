---
page_title: "Upgrading to v1.5.0"
subcategory: ""
description: |-
  Notes on upgrading the Polytomic provider to v1.5.0.
---

# Upgrading to v1.5.0

Version 1.5.0 makes server-managed connection configuration fields read-only. Previously these fields were generated as `Optional` + `Computed`, which silently accepted user-supplied values that the server would overwrite on the next read. They are now `Computed`-only — Terraform will reject configurations that try to set them.

This affects 51 fields across 39 connection types. The values are still readable from state and can be referenced from other resources; only assignment is no longer allowed.

## Migration steps

After upgrading, run `terraform plan`. For any configuration that sets a now-read-only field, Terraform will return an error like:

```
Error: Value Conversion Error
  on main.tf line 12, in resource "polytomic_posthog_connection" "example":
  12:     authenticated_as = "user@example.com"

Can't configure a value for "configuration.authenticated_as": its value
will be decided automatically based on the result of applying this
configuration.
```

Remove the assignment from your configuration and re-run `terraform plan`. No state changes are required.

```hcl
# Before (v1.4.x)
resource "polytomic_posthog_connection" "example" {
  name = "PostHog"
  configuration = {
    api_key          = var.posthog_api_key
    authenticated_as = "user@example.com"  # remove this line
  }
}

# After (v1.5.0)
resource "polytomic_posthog_connection" "example" {
  name = "PostHog"
  configuration = {
    api_key = var.posthog_api_key
  }
}
```

If another resource references the field, the reference continues to work — the value is still computed and exported:

```hcl
output "posthog_user" {
  value = polytomic_posthog_connection.example.configuration.authenticated_as
}
```

## Importer

If you generated your configuration with the `polytomic` importer, regenerate it with the v1.5.0 importer to drop the now-read-only assignments automatically.

## Affected fields

| Connection | Field(s) |
|---|---|
| polytomic_affinity_connection | user |
| polytomic_airtable_connection | connected_user |
| polytomic_attio_connection | workspace_name |
| polytomic_awsathena_connection | aws_user |
| polytomic_awsopensearch_connection | aws_user |
| polytomic_bigquery_connection | client_email, project_id |
| polytomic_customerio_connection | region |
| polytomic_customeriowarehouseexports_connection | aws_user, external_id |
| polytomic_databricks_connection | aws_user, external_id |
| polytomic_dbtprojectrepository_connection | connected_user |
| polytomic_dynamodb_connection | aws_user, external_id |
| polytomic_fbaudience_connection | user_name |
| polytomic_gcs_connection | client_email, project_id |
| polytomic_gmail_connection | user_email |
| polytomic_googleads_connection | connected_user |
| polytomic_googleanalytics_connection | user_email |
| polytomic_googleslides_connection | user_email |
| polytomic_googleworkspace_connection | client_email |
| polytomic_gsheets_connection | user_email |
| polytomic_hubspot_connection | hub_domain, hub_user |
| polytomic_linkedinads_connection | connected_user |
| polytomic_motherduck_connection | aws_user |
| polytomic_msads_connection | username |
| polytomic_outreach_connection | connected_user |
| polytomic_pardot_connection | username |
| polytomic_pinterest_ads_connection | connected_user |
| polytomic_polytomic_metadata_connection | connected_org, connected_user |
| polytomic_posthog_connection | authenticated_as |
| polytomic_redditads_connection | connected_user |
| polytomic_redshift_connection | aws_user, external_id |
| polytomic_redshiftserverless_connection | external_id |
| polytomic_s3_connection | aws_user, external_id |
| polytomic_salesforce_connection | instance_url_override, username |
| polytomic_salesloft_connection | connected_user |
| polytomic_slack_connection | event_url |
| polytomic_tiktok_ads_connection | connected_user |
| polytomic_webhook_connection | secret |
| polytomic_xero_connection | tenant_name, tenant_type |
| polytomic_youtubeanalytics_connection | user_email |
