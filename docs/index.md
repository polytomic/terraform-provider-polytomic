---
page_title: "Polytomic Provider"
subcategory: ""
description: |-
  
---

# Polytomic Provider

The Polytomic provider is used to interact with the resources supported by Polytomic. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources and datasources.



## Example Usage

### API Key 
```terraform
provider "polytomic" {
  api_key     = "<value from settings page>"
}
```

### Deployment API Key (On-Premises only)
```terraform
provider "polytomic" {
  deployment_url     = "polytomic.acmeinc.com"
  deployment_api_key = "<value from deployment environment>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) Polytomic API key
- `deployment_api_key` (String, Sensitive) Polytomic deployment key (required if `api_key` is not set)
- `deployment_url` (String) Polytomic deployment URL (defaults to app.polytomic.com)


## Importing existing resources
Polytomic offers the ability to import existing resources from your Polytomic account into your Terraform state. This allows you to manage existing resources via Terraform. 

### Running the importer
The importer is a separate binary that can be downloaded from the [releases page](https://github.com/polytomic/terraform-provider-polytomic/releases). The following command will run the importer and import all resources into the specified directory.

```bash
./polytomic-importer run --api-key $POLYTOMIC_API_KEY --output $OUTPUT_DIRECTORY
```
