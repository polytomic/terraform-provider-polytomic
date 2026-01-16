package importer

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
	"github.com/polytomic/polytomic-go/bulksync"
	ptclient "github.com/polytomic/polytomic-go/client"
	"github.com/polytomic/terraform-provider-polytomic/provider"
	"github.com/zclconf/go-cty/cty"
)

const (
	BulkSyncResourceFileName = "bulk_syncs.tf"
	BulkSyncResource         = "polytomic_bulk_sync"
)

var (
	_ Importable = &BulkSyncs{}
)

type BulkSyncs struct {
	c *ptclient.Client

	Resources map[string]*polytomic.BulkSyncResponse
}

func NewBulkSyncs(c *ptclient.Client) *BulkSyncs {
	return &BulkSyncs{
		c:         c,
		Resources: make(map[string]*polytomic.BulkSyncResponse),
	}
}

func (b *BulkSyncs) Init(ctx context.Context) error {
	bulkSyncs, err := b.c.BulkSync.List(ctx, &polytomic.BulkSyncListRequest{})
	if err != nil {
		return err
	}
	for _, bulk := range bulkSyncs.Data {
		// Bulk sync names are not unique, so we need to a slug to the name
		// to make it unique.
		name := provider.ValidName(provider.ToSnakeCase(pointer.GetString(bulk.Name)) + "_" + pointer.GetString(bulk.Id)[:8])
		b.Resources[name] = bulk
	}

	return nil
}

func (b *BulkSyncs) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	// Validate schema once before generating any files
	validator, err := NewSchemaValidator(ctx, provider.NewBulkSyncResourceForSchemaIntrospection())
	if err != nil {
		return fmt.Errorf("failed to create schema validator: %w", err)
	}

	for _, name := range sortedKeys(b.Resources) {
		bulkSync := b.Resources[name]
		bulkSchemas, err := b.c.BulkSync.Schemas.List(ctx, pointer.GetString(bulkSync.Id), &bulksync.SchemasListRequest{})
		if err != nil {
			return err
		}
		schemas := make([]string, 0, len(bulkSchemas.Data))
		for _, schema := range bulkSchemas.Data {
			if pointer.GetBool(schema.Enabled) {
				schemas = append(schemas, pointer.GetString(schema.Id))
			}
		}

		// Build the field mapping for this bulk sync
		mapping := b.buildFieldMapping(bulkSync, schemas)

		// Validate the mapping against the actual provider schema
		if err := validator.ValidateMapping(mapping); err != nil {
			return fmt.Errorf("schema validation failed for bulk sync '%s': %w", pointer.GetString(bulkSync.Name), err)
		}

		// Generate HCL file
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()

		resourceBlock := body.AppendNewBlock("resource", []string{BulkSyncResource, name})

		// Basic attributes
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(pointer.GetString(bulkSync.Name)))
		resourceBlock.Body().SetAttributeValue("active", cty.BoolVal(pointer.GetBool(bulkSync.Active)))
		resourceBlock.Body().SetAttributeValue("mode", cty.StringVal(string(pointer.Get(bulkSync.Mode))))

		// Source connection (nested object attribute, not a block)
		// Configuration must be a jsonencoded string
		sourceConfigTokens := wrapJSONEncode(bulkSync.SourceConfiguration) // Wrap entire config
		sourceTokens := hclwrite.Tokens{
			&hclwrite.Token{Bytes: []byte("{\n")},
			&hclwrite.Token{Bytes: []byte("    connection_id = ")},
			&hclwrite.Token{Bytes: []byte(fmt.Sprintf(`"%s"`, pointer.GetString(bulkSync.SourceConnectionId)))},
			&hclwrite.Token{Bytes: []byte("\n")},
			&hclwrite.Token{Bytes: []byte("    configuration = ")},
		}
		sourceTokens = append(sourceTokens, sourceConfigTokens...)
		sourceTokens = append(sourceTokens, hclwrite.Tokens{
			&hclwrite.Token{Bytes: []byte("\n")},
			&hclwrite.Token{Bytes: []byte("  }")},
		}...)
		resourceBlock.Body().SetAttributeRaw("source", sourceTokens)

		// Destination connection (nested object attribute, not a block)
		// Configuration must be a jsonencoded string
		destConfigTokens := wrapJSONEncode(bulkSync.DestinationConfiguration) // Wrap entire config
		destTokens := hclwrite.Tokens{
			&hclwrite.Token{Bytes: []byte("{\n")},
			&hclwrite.Token{Bytes: []byte("    connection_id = ")},
			&hclwrite.Token{Bytes: []byte(fmt.Sprintf(`"%s"`, pointer.GetString(bulkSync.DestinationConnectionId)))},
			&hclwrite.Token{Bytes: []byte("\n")},
			&hclwrite.Token{Bytes: []byte("    configuration = ")},
		}
		destTokens = append(destTokens, destConfigTokens...)
		destTokens = append(destTokens, hclwrite.Tokens{
			&hclwrite.Token{Bytes: []byte("\n")},
			&hclwrite.Token{Bytes: []byte("  }")},
		}...)
		resourceBlock.Body().SetAttributeRaw("destination", destTokens)

		// New fields that replaced "discover"
		if bulkSync.AutomaticallyAddNewFields != nil {
			resourceBlock.Body().SetAttributeValue("automatically_add_new_fields",
				cty.StringVal(string(pointer.Get(bulkSync.AutomaticallyAddNewFields))))
		}
		if bulkSync.AutomaticallyAddNewObjects != nil {
			resourceBlock.Body().SetAttributeValue("automatically_add_new_objects",
				cty.StringVal(string(pointer.Get(bulkSync.AutomaticallyAddNewObjects))))
		}

		// Disable record timestamps if set
		if bulkSync.DisableRecordTimestamps != nil {
			resourceBlock.Body().SetAttributeValue("disable_record_timestamps",
				cty.BoolVal(pointer.GetBool(bulkSync.DisableRecordTimestamps)))
		}

		// Optional: Concurrency limits
		if bulkSync.ConcurrencyLimit != nil {
			resourceBlock.Body().SetAttributeValue("concurrency_limit",
				cty.NumberIntVal(int64(pointer.GetInt(bulkSync.ConcurrencyLimit))))
		}
		if bulkSync.ResyncConcurrencyLimit != nil {
			resourceBlock.Body().SetAttributeValue("resync_concurrency_limit",
				cty.NumberIntVal(int64(pointer.GetInt(bulkSync.ResyncConcurrencyLimit))))
		}

		// Optional: Normalize names
		if bulkSync.NormalizeNames != nil {
			resourceBlock.Body().SetAttributeValue("normalize_names",
				cty.StringVal(string(pointer.Get(bulkSync.NormalizeNames))))
		}

		// Optional: Data cutoff timestamp
		if bulkSync.DataCutoffTimestamp != nil {
			resourceBlock.Body().SetAttributeValue("data_cutoff_timestamp",
				cty.StringVal(bulkSync.DataCutoffTimestamp.Format(time.RFC3339)))
		}

		// Optional: Policies
		if len(bulkSync.Policies) > 0 {
			resourceBlock.Body().SetAttributeValue("policies", typeConverter(bulkSync.Policies))
		}

		// Schemas - convert to array of objects with id and enabled fields
		schemaObjects := make([]map[string]interface{}, 0, len(schemas))
		for _, schemaID := range schemas {
			schemaObjects = append(schemaObjects, map[string]interface{}{
				"id":      schemaID,
				"enabled": true, // If it's in the list, it's enabled
			})
		}
		resourceBlock.Body().SetAttributeValue("schemas", typeConverter(schemaObjects))

		// Schedule
		var schedule map[string]interface{}
		decoder, err := mapstructure.NewDecoder(
			&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  &schedule,
			})
		if err != nil {
			return err
		}
		err = decoder.Decode(bulkSync.Schedule)
		if err != nil {
			return err
		}
		// TODO: @JakeNeyer - multi schedule is not supported
		// add once multi schedule is supported in the provider
		delete(schedule, "multi")
		resourceBlock.Body().SetAttributeValue("schedule", typeConverter(schedule))

		body.AppendNewline()

		writer.Write(ReplaceRefs(hclFile.Bytes(), refs))
	}

	return nil
}

// buildFieldMapping creates a mapping structure for schema validation
// This represents the HCL structure we're generating
func (b *BulkSyncs) buildFieldMapping(bulkSync *polytomic.BulkSyncResponse, schemas []string) map[string]interface{} {
	mapping := map[string]interface{}{
		"name":   pointer.GetString(bulkSync.Name),
		"active": pointer.GetBool(bulkSync.Active),
		"mode":   string(pointer.Get(bulkSync.Mode)),
		"source": map[string]interface{}{
			"connection_id": pointer.GetString(bulkSync.SourceConnectionId),
			"configuration": "{}", // Placeholder for validation
		},
		"destination": map[string]interface{}{
			"connection_id": pointer.GetString(bulkSync.DestinationConnectionId),
			"configuration": "{}", // Placeholder for validation
		},
		"schemas":  schemas,
		"schedule": map[string]interface{}{}, // Placeholder for validation
	}

	// Add optional fields if present
	if bulkSync.AutomaticallyAddNewFields != nil {
		mapping["automatically_add_new_fields"] = string(pointer.Get(bulkSync.AutomaticallyAddNewFields))
	}
	if bulkSync.AutomaticallyAddNewObjects != nil {
		mapping["automatically_add_new_objects"] = string(pointer.Get(bulkSync.AutomaticallyAddNewObjects))
	}
	if bulkSync.DisableRecordTimestamps != nil {
		mapping["disable_record_timestamps"] = pointer.GetBool(bulkSync.DisableRecordTimestamps)
	}
	if bulkSync.ConcurrencyLimit != nil {
		mapping["concurrency_limit"] = pointer.GetInt(bulkSync.ConcurrencyLimit)
	}
	if bulkSync.ResyncConcurrencyLimit != nil {
		mapping["resync_concurrency_limit"] = pointer.GetInt(bulkSync.ResyncConcurrencyLimit)
	}
	if bulkSync.NormalizeNames != nil {
		mapping["normalize_names"] = string(pointer.Get(bulkSync.NormalizeNames))
	}
	if bulkSync.DataCutoffTimestamp != nil {
		mapping["data_cutoff_timestamp"] = bulkSync.DataCutoffTimestamp.Format(time.RFC3339)
	}
	if len(bulkSync.Policies) > 0 {
		mapping["policies"] = bulkSync.Policies
	}

	return mapping
}

func (b *BulkSyncs) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, name := range sortedKeys(b.Resources) {
		bulkSync := b.Resources[name]
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			BulkSyncResource,
			name,
			pointer.GetString(bulkSync.Id))))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", pointer.GetString(bulkSync.Name))))
	}
	return nil
}

func (b *BulkSyncs) Filename() string {
	return BulkSyncResourceFileName
}

func (b *BulkSyncs) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, bulk := range b.Resources {
		result[pointer.GetString(bulk.Id)] = fmt.Sprintf("%s.%s.id", BulkSyncResource, name)
	}
	return result
}

func (b *BulkSyncs) DatasourceRefs() map[string]string {
	return nil
}

func (b *BulkSyncs) Variables() []Variable {
	return nil
}
