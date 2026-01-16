package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// actorAttributes returns the schema attributes for a CommonOutputActor object.
// This is used for created_by and updated_by fields across multiple resources.
func actorAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Actor ID",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Actor name",
			Computed:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Actor type (user, system, organization, partner)",
			Computed:            true,
		},
	}
}

// actorAttrTypes returns the attribute types map for a CommonOutputActor object.
// This is used when converting SDK actor objects to Terraform object values.
func actorAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
		"type": types.StringType,
	}
}
