package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go/v25"
	ptclient "github.com/polytomic/polytomic-go/v25/client"
	ptcore "github.com/polytomic/polytomic-go/v25/core"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

var errSchemaNotFound = errors.New("schema not found")

// schemaFieldModel is a connection schema field as reported by the connection
// schema data sources.
type schemaFieldModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Type               types.String `tfsdk:"type"`
	TypeSpec           types.String `tfsdk:"type_spec"`
	RemoteType         types.String `tfsdk:"remote_type"`
	Path               types.String `tfsdk:"path"`
	UserManaged        types.Bool   `tfsdk:"user_managed"`
	IsPrimaryKey       types.Bool   `tfsdk:"is_primary_key"`
	SourcePrimaryKey   types.Bool   `tfsdk:"source_primary_key"`
	PrimaryKeyOverride types.Bool   `tfsdk:"primary_key_override"`
}

var schemaFieldObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":                   types.StringType,
		"name":                 types.StringType,
		"type":                 types.StringType,
		"type_spec":            types.StringType,
		"remote_type":          types.StringType,
		"path":                 types.StringType,
		"user_managed":         types.BoolType,
		"is_primary_key":       types.BoolType,
		"source_primary_key":   types.BoolType,
		"primary_key_override": types.BoolType,
	},
}

func schemaFieldDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Field ID",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Field name, including any label override",
			Computed:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Field type, including any type override",
			Computed:            true,
		},
		"type_spec": schema.StringAttribute{
			MarkdownDescription: "Detailed type specification, JSON encoded; null when the source does not report one",
			Computed:            true,
		},
		"remote_type": schema.StringAttribute{
			MarkdownDescription: "Type of the field in the source system",
			Computed:            true,
		},
		"path": schema.StringAttribute{
			MarkdownDescription: "JSONPath used to extract the field from each source record; only set on document-style sources",
			Computed:            true,
		},
		"user_managed": schema.BoolAttribute{
			MarkdownDescription: "Whether the field's definition comes from a user-defined field or override, such as `polytomic_connection_schema_field`",
			Computed:            true,
		},
		"is_primary_key": schema.BoolAttribute{
			MarkdownDescription: "Whether the field is part of the schema's primary key, including any override",
			Computed:            true,
		},
		"source_primary_key": schema.BoolAttribute{
			MarkdownDescription: "Whether the source reports the field as part of the schema's primary key; null when the Polytomic deployment does not report it",
			Computed:            true,
		},
		"primary_key_override": schema.BoolAttribute{
			MarkdownDescription: "Primary key status set by an override, such as `polytomic_connection_schema_primary_keys`; null when the field has no override",
			Computed:            true,
		},
	}
}

func newSchemaFieldModel(f *polytomic.SchemaField) (schemaFieldModel, error) {
	m := schemaFieldModel{
		ID:                 types.StringPointerValue(f.ID),
		Name:               types.StringPointerValue(f.Name),
		Type:               types.StringNull(),
		TypeSpec:           types.StringNull(),
		RemoteType:         types.StringPointerValue(f.RemoteType),
		Path:               types.StringPointerValue(f.Path),
		UserManaged:        types.BoolValue(pointer.GetBool(f.UserManaged)),
		IsPrimaryKey:       types.BoolValue(pointer.GetBool(f.IsPrimaryKey)),
		SourcePrimaryKey:   types.BoolPointerValue(f.SourcePrimaryKey),
		PrimaryKeyOverride: types.BoolPointerValue(f.PrimaryKeyOverride),
	}
	if f.Type != nil {
		m.Type = types.StringValue(string(*f.Type))
	}
	if f.TypeSpec != nil {
		spec, err := json.Marshal(f.TypeSpec)
		if err != nil {
			return m, fmt.Errorf("encoding type_spec for field %s: %w", pointer.GetString(f.ID), err)
		}
		m.TypeSpec = types.StringValue(string(spec))
	}
	return m, nil
}

func newSchemaFieldModels(fields []*polytomic.SchemaField) ([]schemaFieldModel, error) {
	models := make([]schemaFieldModel, 0, len(fields))
	for _, f := range fields {
		if f == nil {
			continue
		}
		m, err := newSchemaFieldModel(f)
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

// fetchSchema returns a connection schema, or errSchemaNotFound when the
// connection or schema does not exist. On a new connection, it waits for the
// first schema inspection before reporting a schema missing.
func fetchSchema(ctx context.Context, client *ptclient.Client, connectionID, schemaID string) (*polytomic.Schema, error) {
	schemaData, err := retryUntilSchemaCached(ctx, client, connectionID, func() (*polytomic.Schema, error) {
		resp, err := client.Schemas.Get(ctx, connectionID, schemaID)
		if err != nil {
			return nil, err
		}
		return resp.Data, nil
	})
	if err != nil {
		if isNotFound(err) {
			return nil, errSchemaNotFound
		}
		return nil, err
	}
	if schemaData == nil {
		return nil, errors.New("API returned nil schema data")
	}
	return schemaData, nil
}

// Variables so tests can shorten them.
var (
	schemaCacheTimeout  = 5 * time.Minute
	schemaCacheInterval = 2 * time.Second
)

// retryUntilSchemaCached calls fn again while it returns 404 and the
// connection's schema cache has not finished its first refresh. The API
// inspects a new connection's schemas in the background and reports every
// schema missing until then. The 404 is returned as is once the cache has been
// refreshed, when the cache status can't be read (for example, because the
// connection is gone), or after schemaCacheTimeout.
func retryUntilSchemaCached[T any](ctx context.Context, client *ptclient.Client, connectionID string, fn func() (T, error)) (T, error) {
	deadline := time.Now().Add(schemaCacheTimeout)
	cached := false
	for {
		result, err := fn()
		if !isNotFound(err) || cached {
			return result, err
		}
		status, statusErr := client.Schemas.GetStatus(ctx, connectionID)
		if statusErr != nil {
			return result, err
		}
		if status.GetData().GetLastRefreshFinished() != nil {
			// Call fn once more, in case the refresh finished after it ran.
			cached = true
			continue
		}
		if time.Now().After(deadline) {
			return result, err
		}
		tflog.Debug(ctx, "Waiting for the connection's first schema inspection", map[string]any{
			"connection_id": connectionID,
			"cache_status":  pointer.GetString(status.GetData().GetCacheStatus()),
		})
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(schemaCacheInterval):
		}
	}
}

func findSchemaField(s *polytomic.Schema, fieldID string) *polytomic.SchemaField {
	for _, f := range s.Fields {
		if f != nil && pointer.GetString(f.ID) == fieldID {
			return f
		}
	}
	return nil
}

func isNotFound(err error) bool {
	pErr := &ptcore.APIError{}
	return errors.As(err, &pErr) && pErr.StatusCode == http.StatusNotFound
}

// connectionOrganization returns the ID of the organization that owns a
// connection, or "" when it cannot be determined.
func connectionOrganization(ctx context.Context, client *ptclient.Client, connectionID string) string {
	conn, err := client.Connections.Get(ctx, connectionID)
	if err != nil || conn.Data == nil {
		return ""
	}
	return pointer.GetString(conn.Data.OrganizationID)
}

// resourceOrganization returns the organization to record for a resource on a
// connection: the configured one, else the connection's, else
// providerclient.DefaultOrganization.
func resourceOrganization(ctx context.Context, client *ptclient.Client, org types.String, connectionID string) string {
	if !org.IsNull() && !org.IsUnknown() && org.ValueString() != "" {
		return org.ValueString()
	}
	if id := connectionOrganization(ctx, client, connectionID); id != "" {
		return id
	}
	return providerclient.DefaultOrganization
}
