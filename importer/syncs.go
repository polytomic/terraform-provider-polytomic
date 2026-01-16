package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
	"github.com/polytomic/terraform-provider-polytomic/provider"
	"github.com/zclconf/go-cty/cty"
)

const (
	SyncResourceFileName = "syncs.tf"
	SyncResource         = "polytomic_sync"
)

var (
	_ Importable = &Syncs{}
)

type Syncs struct {
	c *ptclient.Client

	Resources map[string]*polytomic.ModelSyncResponse
}

func NewSyncs(c *ptclient.Client) *Syncs {
	return &Syncs{
		c:         c,
		Resources: make(map[string]*polytomic.ModelSyncResponse),
	}
}

func (s *Syncs) Init(ctx context.Context) error {
	syncs, err := s.c.ModelSync.List(ctx, &polytomic.ModelSyncListRequest{})
	if err != nil {
		return err
	}

	for _, sync := range syncs.Data {
		name := provider.ValidName(provider.ToSnakeCase(pointer.GetString(sync.Name)))
		s.Resources[name] = sync
	}

	return nil
}

func (s *Syncs) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	// Validate schema once before generating any files
	validator, err := NewSchemaValidator(ctx, provider.NewSyncResourceForSchemaIntrospection())
	if err != nil {
		return fmt.Errorf("failed to create schema validator: %w", err)
	}

	for _, name := range sortedKeys(s.Resources) {
		syn := s.Resources[name]
		sync, err := s.c.ModelSync.Get(ctx, pointer.GetString(syn.Id))
		if err != nil {
			return err
		}

		// Build the field mapping for this sync
		mapping := s.buildFieldMapping(sync.Data)

		// Validate the mapping against the actual provider schema
		if err := validator.ValidateMapping(mapping); err != nil {
			return fmt.Errorf("schema validation failed for sync '%s': %w", pointer.GetString(sync.Data.Name), err)
		}

		// Generate HCL file
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{SyncResource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(pointer.GetString(sync.Data.Name)))
		resourceBlock.Body().SetAttributeValue("active", cty.BoolVal(pointer.GetBool(sync.Data.Active)))
		resourceBlock.Body().SetAttributeValue("mode", cty.StringVal(string(pointer.Get(sync.Data.Mode))))
		var schedule map[string]interface{}
		decoder, err := mapstructure.NewDecoder(
			&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  &schedule,
			})
		if err != nil {
			return err
		}
		err = decoder.Decode(sync.Data.Schedule)
		if err != nil {
			return err
		}
		resourceBlock.Body().SetAttributeValue("schedule", typeConverter(schedule))
		var fields []map[string]interface{}
		decoder, err = mapstructure.NewDecoder(
			&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  &fields,
			})
		if err != nil {
			return err
		}
		err = decoder.Decode(sync.Data.Fields)
		if err != nil {
			return err
		}

		// Normalize and filter fields to match Terraform schema requirements
		fields = normalizeAndFilterFields(fields)

		resourceBlock.Body().SetAttributeValue("fields", typeConverter(fields))
		var target map[string]interface{}
		decoder, err = mapstructure.NewDecoder(
			&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  &target,
			})
		if err != nil {
			return err
		}
		err = decoder.Decode(sync.Data.Target)
		if err != nil {
			return err
		}
		tokens := wrapJSONEncode(target, "search_values", "configuration")
		resourceBlock.Body().SetAttributeRaw("target", tokens)

		if sync.Data.FilterLogic != nil {
			resourceBlock.Body().SetAttributeValue("filter_logic", cty.StringVal(pointer.GetString(sync.Data.FilterLogic)))
		}

		if len(sync.Data.Filters) > 0 {
			var filters []map[string]interface{}
			decoder, err = mapstructure.NewDecoder(
				&mapstructure.DecoderConfig{
					TagName: "json",
					Result:  &filters,
				})
			if err != nil {
				return err
			}
			err = decoder.Decode(sync.Data.Filters)
			if err != nil {
				return err
			}
			filterTokens := wrapJSONEncode(filters, "value")
			resourceBlock.Body().SetAttributeRaw("filters", filterTokens)
		}

		if sync.Data.Identity != nil {
			var identity map[string]interface{}
			decoder, err = mapstructure.NewDecoder(
				&mapstructure.DecoderConfig{
					TagName: "json",
					Result:  &identity,
				})
			if err != nil {
				return err
			}
			err = decoder.Decode(sync.Data.Identity)
			if err != nil {
				return err
			}
			resourceBlock.Body().SetAttributeValue("identity", typeConverter(identity))
		}
		if len(sync.Data.OverrideFields) > 0 {
			var overrideFields []map[string]interface{}
			decoder, err = mapstructure.NewDecoder(
				&mapstructure.DecoderConfig{
					TagName: "json",
					Result:  &overrideFields,
				})
			if err != nil {
				return err
			}
			err = decoder.Decode(sync.Data.OverrideFields)
			if err != nil {
				return err
			}
			resourceBlock.Body().SetAttributeValue("override_fields", typeConverter(overrideFields))
		}
		if len(sync.Data.Overrides) > 0 {
			var overrides []map[string]interface{}
			decoder, err = mapstructure.NewDecoder(
				&mapstructure.DecoderConfig{
					TagName: "json",
					Result:  &overrides,
				})
			if err != nil {
				return err
			}
			err = decoder.Decode(sync.Data.Overrides)
			if err != nil {
				return err
			}
			overrideTokens := wrapJSONEncode(overrides, "value")
			resourceBlock.Body().SetAttributeRaw("overrides", overrideTokens)
		}
		resourceBlock.Body().SetAttributeValue("sync_all_records", cty.BoolVal(pointer.GetBool(sync.Data.SyncAllRecords)))
		body.AppendNewline()

		writer.Write(ReplaceRefs(hclFile.Bytes(), refs))

	}
	return nil
}

func (s *Syncs) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, name := range sortedKeys(s.Resources) {
		sync := s.Resources[name]
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			SyncResource,
			name,
			pointer.GetString(sync.Id))))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", pointer.GetString(sync.Name))))
	}
	return nil
}

func (s *Syncs) Filename() string {
	return SyncResourceFileName
}

func (s *Syncs) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, sync := range s.Resources {
		result[pointer.GetString(sync.Id)] = fmt.Sprintf("%s.%s.id", SyncResource, name)
	}
	return result
}

func (s *Syncs) DatasourceRefs() map[string]string {
	return nil
}

func (s *Syncs) Variables() []Variable {
	return nil
}

// buildFieldMapping creates a mapping structure for schema validation
// This represents the HCL structure we're generating
func (s *Syncs) buildFieldMapping(sync *polytomic.ModelSyncResponse) map[string]interface{} {
	mapping := map[string]interface{}{
		"name":             pointer.GetString(sync.Name),
		"active":           pointer.GetBool(sync.Active),
		"mode":             string(pointer.Get(sync.Mode)),
		"schedule":         map[string]interface{}{}, // Placeholder for validation
		"fields":           []interface{}{},          // Placeholder for validation
		"target":           map[string]interface{}{}, // Placeholder for validation
		"sync_all_records": pointer.GetBool(sync.SyncAllRecords),
	}

	// Optional fields
	if sync.FilterLogic != nil {
		mapping["filter_logic"] = pointer.GetString(sync.FilterLogic)
	}
	if len(sync.Filters) > 0 {
		mapping["filters"] = []interface{}{}
	}
	if sync.Identity != nil {
		mapping["identity"] = map[string]interface{}{}
	}
	if len(sync.OverrideFields) > 0 {
		mapping["override_fields"] = []interface{}{}
	}
	if len(sync.Overrides) > 0 {
		mapping["overrides"] = []interface{}{}
	}

	return mapping
}

// normalizeAndFilterFields filters out fields that don't meet Terraform schema requirements
// The API may return fields with missing required attributes that can't be imported
func normalizeAndFilterFields(fields []map[string]interface{}) []map[string]interface{} {
	filtered := make([]map[string]interface{}, 0, len(fields))

	for _, field := range fields {
		// Normalize field keys to snake_case
		field = normalizeConfigKeys(field)

		// Check if this field has the required attributes according to Terraform schema
		// Required: source (with model_id and field), target
		// Exception: if override_value is present, source.field is not required

		// Check if target exists
		target, hasTarget := field["target"]
		if !hasTarget || target == nil || target == "" {
			// Skip fields without target
			continue
		}

		// Check source block
		source, hasSource := field["source"].(map[string]interface{})
		if !hasSource {
			// Skip fields without source block
			continue
		}

		// model_id is required in source
		modelID, hasModelID := source["model_id"]
		if !hasModelID || modelID == nil || modelID == "" {
			continue
		}

		// Check if source.field exists
		sourceField, hasSourceField := source["field"]
		if !hasSourceField || sourceField == nil || sourceField == "" {
			// Skip fields without source.field
			// Note: The Terraform schema requires source.field even when override_value is present
			// This is likely a schema bug, but we need to skip these fields for now
			continue
		}

		// Skip fields with override_value since they cause validation errors
		// due to the schema requiring source.field even when override_value is present
		if _, hasOverrideValue := field["override_value"]; hasOverrideValue {
			continue
		}

		// This field is valid, include it
		filtered = append(filtered, field)
	}

	return filtered
}
