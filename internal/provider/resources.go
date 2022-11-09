package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	resourceList = []func() resource.Resource{
		func() resource.Resource { return &organizationResource{} },
		func() resource.Resource { return &userResource{} },
	}
)

func resources() []func() resource.Resource {
	all := append(connection_resources, resourceList...)

	return all
}
