package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.ResourceType = athenaConnectionResourceType{}
var _ resource.Resource = athenaConnectionResource{}

var _ resource.ResourceWithImportState = athenaConnectionResource{}

type athenaConnectionResourceType struct{}

func (t athenaConnectionResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "AWS Athena Connection",
		Attributes: map[string]tfsdk.Attribute{
			"organization": {
				MarkdownDescription: "Organization ID",
				Required:            true,
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
					"access_key_id": {
						Type:      types.StringType,
						Required:  true,
						Sensitive: true,
					},
					"access_key_secret": {
						Type:      types.StringType,
						Required:  true,
						Sensitive: true,
					},
					"region": {
						Type:     types.StringType,
						Required: true,
					},
					"output_bucket": {
						MarkdownDescription: "S3 bucket for output storage, with optional prefix. Examples: `bucket-name`, `bucket-name/prefix`.",
						Type:                types.StringType,
						Required:            true,
					},
				}),
				Required: true,
			},

			"id": {
				Computed:            true,
				MarkdownDescription: "AWS Athena Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t athenaConnectionResourceType) NewResource(ctx context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return athenaConnectionResource{
		provider: provider,
	}, diags
}

type athenaConnectionResource struct {
	provider ptProvider
}

func (r athenaConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.provider.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.Value,
			Type:           "awsathena",
			OrganizationId: data.Organization.Value,
			Configuration: polytomic.AthenaConfiguration{
				AccessKeyID:     data.Configuration.Attrs["access_key_id"].(types.String).Value,
				AccessKeySecret: data.Configuration.Attrs["access_key_secret"].(types.String).Value,
				Region:          data.Configuration.Attrs["region"].(types.String).Value,
				OutputBucket:    data.Configuration.Attrs["output_bucket"].(types.String).Value,
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.String{Value: created.ID}
	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "awsathena"})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r athenaConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.provider.client.Connections().Get(ctx, uuid.MustParse(data.Id.Value))
	if err != nil {
		if err.Error() == ConnectionNotFoundErr {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading connection: %s", err))
		return
	}

	data.Id = types.String{Value: connection.ID}
	data.Organization = types.String{Value: connection.OrganizationId}
	data.Name = types.String{Value: connection.Name}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r athenaConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.provider.client.Connections().Update(ctx,
		uuid.MustParse(data.Id.Value),
		polytomic.UpdateConnectionMutation{
			Name:           data.Name.Value,
			OrganizationId: data.Organization.Value,
			Configuration: polytomic.AthenaConfiguration{
				AccessKeyID:     data.Configuration.Attrs["access_key_id"].(types.String).Value,
				AccessKeySecret: data.Configuration.Attrs["access_key_secret"].(types.String).Value,
				Region:          data.Configuration.Attrs["region"].(types.String).Value,
				OutputBucket:    data.Configuration.Attrs["output_bucket"].(types.String).Value,
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.String{Value: updated.ID}
	data.Organization = types.String{Value: updated.OrganizationId}
	data.Name = types.String{Value: updated.Name}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r athenaConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.Connections().Delete(ctx, uuid.MustParse(data.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
		return
	}
}

func (r athenaConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
