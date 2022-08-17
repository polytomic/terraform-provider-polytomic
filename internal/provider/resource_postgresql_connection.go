package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
)

var postgresqlResourceSchema = tfsdk.Schema{
	MarkdownDescription: "Postgresql Connection",
	Attributes: map[string]tfsdk.Attribute{
		"organization": {
			MarkdownDescription: "Organization ID",
			Optional:            true,
			Type:                types.StringType,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.RequiresReplace(),
			},
		},
		"name": {
			Type:     types.StringType,
			Required: true,
		},
		"configuration": {
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"hostname": {
					Type:     types.StringType,
					Required: true,
				},
				"username": {
					Type:     types.StringType,
					Required: true,
				},
				"password": {
					Type:     types.StringType,
					Required: true,
				},
				"database": {
					Type:     types.StringType,
					Required: true,
				},
				"port": {
					Type:     types.Int64Type,
					Required: true,
				},
				"ssl": {
					Type:     types.BoolType,
					Optional: true,
				},
			}),
			Required: true,
		},

		"id": {
			Computed:            true,
			MarkdownDescription: "Postgresql Connection identifier",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
	},
}

func getConfiguration(data connectionResourceData) (interface{}, error) {
	return polytomic.PostgresqlConfiguration{
		Hostname: data.Configuration.Attrs["hostname"].(types.String).Value,
		Username: data.Configuration.Attrs["username"].(types.String).Value,
		Password: data.Configuration.Attrs["password"].(types.String).Value,
		Database: data.Configuration.Attrs["database"].(types.String).Value,
		Port:     int(data.Configuration.Attrs["port"].(types.Int64).Value),
		SSL:      data.Configuration.Attrs["ssl"].(types.Bool).Value,
	}, nil
}
