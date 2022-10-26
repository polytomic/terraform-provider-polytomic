package provider

import "github.com/hashicorp/terraform-plugin-framework/provider"

var (
	resourceList = map[string]provider.ResourceType{
		"polytomic_organization": organizationResourceType{},
		"polytomic_user":         userResourceType{},
	}
)

func resources() map[string]provider.ResourceType {
	all := map[string]provider.ResourceType{}
	for k, v := range resourceList {
		all[k] = v
	}
	for k, v := range connection_resources {
		all[k] = v
	}
	return all
}
