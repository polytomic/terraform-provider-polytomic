package importer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/polytomic/polytomic-go/v25"
	ptclient "github.com/polytomic/polytomic-go/v25/client"
	ptcore "github.com/polytomic/polytomic-go/v25/core"
	"github.com/polytomic/terraform-provider-polytomic/provider"
	"github.com/rs/zerolog/log"
	"github.com/zclconf/go-cty/cty"
)

const (
	SchemaOverridesFileName   = "schema_overrides.tf"
	SchemaFieldResource       = "polytomic_connection_schema_field"
	SchemaPrimaryKeysResource = "polytomic_connection_schema_primary_keys"
)

var (
	_ Importable = &SchemaOverrides{}
)

// SchemaOverrides imports the fields and primary keys users have defined or
// overridden on connection schemas.
type SchemaOverrides struct {
	c              *ptclient.Client
	organizationID string

	Fields      map[string]schemaFieldOverride
	PrimaryKeys map[string]schemaPrimaryKeys
}

type schemaFieldOverride struct {
	ConnectionID string
	SchemaID     string
	Field        *polytomic.SchemaField
}

type schemaPrimaryKeys struct {
	ConnectionID string
	SchemaID     string
	FieldIDs     []string
}

func NewSchemaOverrides(c *ptclient.Client, organizationID string) *SchemaOverrides {
	return &SchemaOverrides{
		c:              c,
		organizationID: organizationID,
		Fields:         make(map[string]schemaFieldOverride),
		PrimaryKeys:    make(map[string]schemaPrimaryKeys),
	}
}

func (s *SchemaOverrides) Init(ctx context.Context) error {
	conns, err := s.c.Connections.List(ctx)
	if err != nil {
		return err
	}
	for _, conn := range conns.Data {
		connectionID := pointer.GetString(conn.ID)
		source, err := s.c.BulkSync.GetSource(ctx, connectionID, &polytomic.BulkSyncGetSourceRequest{
			IncludeFields: pointer.To(true),
		})
		if err != nil {
			if isClientError(err) {
				// Connections that cannot be read from have no schemas.
				log.Debug().Str("connection_id", connectionID).AnErr("error", err).Msg("skipping connection schemas")
				continue
			}
			return fmt.Errorf("listing schemas for connection %s: %w", connectionID, err)
		}
		if source.Data == nil {
			continue
		}
		if !s.add(pointer.GetString(conn.Name), connectionID, source.Data.Schemas) {
			log.Warn().Str("connection_id", connectionID).
				Msg("skipping primary key overrides: the Polytomic deployment does not report them")
		}
	}
	return nil
}

// add records the overrides on one connection's schemas. It returns false when
// the deployment does not report primary key overrides, so none could be
// recorded.
func (s *SchemaOverrides) add(connectionName, connectionID string, schemas []*polytomic.Schema) bool {
	sawFields, reportsOverrides := false, false
	for _, sch := range schemas {
		for _, f := range sch.GetFields() {
			sawFields = true
			if f != nil && f.SourcePrimaryKey != nil {
				reportsOverrides = true
			}
		}
	}

	for _, sch := range schemas {
		if sch == nil {
			continue
		}
		schemaID := pointer.GetString(sch.ID)
		keys := []string{}
		overridden := false
		for _, f := range sch.Fields {
			if f == nil {
				continue
			}
			fieldID := pointer.GetString(f.ID)
			if pointer.GetBool(f.UserManaged) {
				s.Fields[uniqueName(s.Fields, connectionName, schemaID, fieldID)] = schemaFieldOverride{
					ConnectionID: connectionID,
					SchemaID:     schemaID,
					Field:        f,
				}
			}
			if pointer.GetBool(f.IsPrimaryKey) {
				keys = append(keys, fieldID)
			}
			if f.PrimaryKeyOverride != nil {
				overridden = true
			}
		}

		if !overridden {
			continue
		}
		if len(keys) == 0 {
			log.Warn().Str("connection_id", connectionID).Str("schema_id", schemaID).
				Msg("skipping primary key overrides that leave the schema without a primary key")
			continue
		}
		sort.Strings(keys)
		s.PrimaryKeys[uniqueName(s.PrimaryKeys, connectionName, schemaID)] = schemaPrimaryKeys{
			ConnectionID: connectionID,
			SchemaID:     schemaID,
			FieldIDs:     keys,
		}
	}

	return !sawFields || reportsOverrides
}

func (s *SchemaOverrides) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	pkValidator, err := NewSchemaValidator(ctx, provider.NewConnectionSchemaPrimaryKeysResource())
	if err != nil {
		return fmt.Errorf("failed to create schema validator: %w", err)
	}
	fieldValidator, err := NewSchemaValidator(ctx, provider.NewConnectionSchemaFieldResource())
	if err != nil {
		return fmt.Errorf("failed to create schema validator: %w", err)
	}

	for _, name := range sortedKeys(s.PrimaryKeys) {
		pk := s.PrimaryKeys[name]
		err := pkValidator.ValidateMapping(map[string]interface{}{
			"connection_id": pk.ConnectionID,
			"schema_id":     pk.SchemaID,
			"field_ids":     pk.FieldIDs,
		})
		if err != nil {
			return fmt.Errorf("schema validation failed for primary keys of schema '%s': %w", pk.SchemaID, err)
		}

		fieldIDs := make([]cty.Value, len(pk.FieldIDs))
		for i, id := range pk.FieldIDs {
			fieldIDs[i] = cty.StringVal(id)
		}

		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		block := body.AppendNewBlock("resource", []string{SchemaPrimaryKeysResource, name}).Body()
		block.SetAttributeValue("connection_id", cty.StringVal(pk.ConnectionID))
		setOrganizationLocal(block)
		block.SetAttributeValue("schema_id", cty.StringVal(pk.SchemaID))
		block.SetAttributeValue("field_ids", cty.ListVal(fieldIDs))
		body.AppendNewline()

		if _, err := writer.Write(hclwrite.Format(ReplaceRefs(hclFile.Bytes(), refs))); err != nil {
			return err
		}
	}

	for _, name := range sortedKeys(s.Fields) {
		o := s.Fields[name]
		f := o.Field
		mapping := map[string]interface{}{
			"connection_id": o.ConnectionID,
			"schema_id":     o.SchemaID,
			"field_id":      pointer.GetString(f.ID),
		}
		if pointer.GetString(f.Name) != "" {
			mapping["label"] = pointer.GetString(f.Name)
		}
		typeName, precision, scale, typeSpec, err := provider.SchemaFieldTypeAttributes(f)
		if err != nil {
			return err
		}
		if typeName != "" {
			mapping["type"] = typeName
		}
		if precision != nil {
			mapping["precision"] = *precision
		}
		if scale != nil {
			mapping["scale"] = *scale
		}
		if typeSpec != "" {
			mapping["type_spec"] = typeSpec
		}
		if pointer.GetString(f.Path) != "" {
			mapping["path"] = pointer.GetString(f.Path)
		}
		if err := fieldValidator.ValidateMapping(mapping); err != nil {
			return fmt.Errorf("schema validation failed for field '%s' of schema '%s': %w", pointer.GetString(f.ID), o.SchemaID, err)
		}

		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		block := body.AppendNewBlock("resource", []string{SchemaFieldResource, name}).Body()
		block.SetAttributeValue("connection_id", cty.StringVal(o.ConnectionID))
		setOrganizationLocal(block)
		block.SetAttributeValue("schema_id", cty.StringVal(o.SchemaID))
		block.SetAttributeValue("field_id", cty.StringVal(pointer.GetString(f.ID)))
		for _, attr := range []string{"label", "type", "path"} {
			if v, ok := mapping[attr]; ok {
				block.SetAttributeValue(attr, cty.StringVal(v.(string)))
			}
		}
		for _, attr := range []string{"precision", "scale"} {
			if v, ok := mapping[attr]; ok {
				block.SetAttributeValue(attr, cty.NumberIntVal(v.(int64)))
			}
		}
		if typeSpec != "" {
			tokens, err := jsonEncodeTokens([]byte(typeSpec))
			if err != nil {
				return fmt.Errorf("encoding type_spec for field '%s' of schema '%s': %w", pointer.GetString(f.ID), o.SchemaID, err)
			}
			block.SetAttributeRaw("type_spec", tokens)
		}
		body.AppendNewline()

		if _, err := writer.Write(hclwrite.Format(ReplaceRefs(hclFile.Bytes(), refs))); err != nil {
			return err
		}
	}

	return nil
}

func (s *SchemaOverrides) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, name := range sortedKeys(s.PrimaryKeys) {
		pk := s.PrimaryKeys[name]
		fmt.Fprintf(writer, "terraform import %s.%s %s\n",
			SchemaPrimaryKeysResource, name,
			shellQuote(strings.Join([]string{s.organizationID, pk.ConnectionID, pk.SchemaID}, "/")))
	}
	for _, name := range sortedKeys(s.Fields) {
		o := s.Fields[name]
		fmt.Fprintf(writer, "terraform import %s.%s %s\n",
			SchemaFieldResource, name,
			shellQuote(provider.SchemaFieldResourceID(s.organizationID, o.ConnectionID, o.SchemaID, pointer.GetString(o.Field.ID))))
	}
	return nil
}

func (s *SchemaOverrides) Filename() string {
	return SchemaOverridesFileName
}

func (s *SchemaOverrides) ResourceRefs() map[string]string {
	return nil
}

func (s *SchemaOverrides) DatasourceRefs() map[string]string {
	return nil
}

func (s *SchemaOverrides) Variables() []Variable {
	return nil
}

func setOrganizationLocal(body *hclwrite.Body) {
	body.SetAttributeTraversal("organization", hcl.Traversal{
		hcl.TraverseRoot{Name: "local"},
		hcl.TraverseAttr{Name: "organization_id"},
	})
}

// uniqueName builds a resource name from parts, suffixing it when the name is
// already taken.
func uniqueName[V any](taken map[string]V, parts ...string) string {
	base := provider.ValidName(provider.ToSnakeCase(strings.Join(parts, "_")))
	name := base
	for i := 2; ; i++ {
		if _, ok := taken[name]; !ok {
			return name
		}
		name = fmt.Sprintf("%s_%d", base, i)
	}
}

func isClientError(err error) bool {
	apiErr := &ptcore.APIError{}
	return errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500
}

// shellQuote quotes s for import.sh when it contains characters the shell
// would interpret.
func shellQuote(s string) string {
	safe := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("-_./:@", r)
	}
	if strings.IndexFunc(s, func(r rune) bool { return !safe(r) }) < 0 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
