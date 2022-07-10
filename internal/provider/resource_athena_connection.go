package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = athenaConnectionResourceType{}
var _ tfsdk.Resource = athenaConnectionResource{}

var _ tfsdk.ResourceWithImportState = athenaConnectionResource{}

type athenaConnectionResourceType struct{}

func (t athenaConnectionResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "AWS Athena Connection",
		Attributes: map[string]tfsdk.Attribute{
			"workspace": {
				MarkdownDescription: "Workspace ID",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
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
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t athenaConnectionResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return athenaConnectionResource{
		provider: provider,
	}, diags
}

type athenaConfigurationData struct {
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	AccessKeySecret types.String `tfsdk:"access_key_secret"`
	Region          types.String `tfsdk:"region"`
	OutputBucket    types.String `tfsdk:"output_bucket"`
}

type athenaConnectionResource struct {
	provider provider
}

func (r athenaConnectionResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data connectionResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.provider.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:        data.Name.Value,
			Type:        "awsathena",
			WorkspaceId: data.Workspace.Value,
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

func (r athenaConnectionResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data connectionResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.provider.client.Connections().Get(ctx, uuid.MustParse(data.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading connection: %s", err))
		return
	}

	data.Id = types.String{Value: connection.ID}
	data.Workspace = types.String{Value: connection.WorkspaceID}
	data.Name = types.String{Value: connection.Name}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r athenaConnectionResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data connectionResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.provider.client.Connections().Update(ctx,
		uuid.MustParse(data.Id.Value),
		polytomic.UpdateConnectionMutation{
			Name:        data.Name.Value,
			WorkspaceId: data.Workspace.Value,
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
	data.Workspace = types.String{Value: updated.WorkspaceID}
	data.Name = types.String{Value: updated.Name}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r athenaConnectionResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
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

func (r athenaConnectionResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
