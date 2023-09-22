package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
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
	c *polytomic.Client

	Resources map[string]polytomic.SyncResponse
}

func NewSyncs(c *polytomic.Client) *Syncs {
	return &Syncs{
		c:         c,
		Resources: make(map[string]polytomic.SyncResponse),
	}
}

func (s *Syncs) Init(ctx context.Context) error {
	syncs, err := s.c.Syncs().List(ctx)
	if err != nil {
		return err
	}

	for _, sync := range syncs {
		name := provider.ValidName(provider.ToSnakeCase(sync.Name))
		s.Resources[name] = sync
	}

	return nil
}

func (s *Syncs) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	for _, name := range sortedKeys(s.Resources) {
		syn := s.Resources[name]
		sync, err := s.c.Syncs().Get(ctx, syn.ID)
		if err != nil {
			return err
		}
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{SyncResource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(sync.Name))
		resourceBlock.Body().SetAttributeValue("mode", cty.StringVal(sync.Mode))
		var schedule map[string]interface{}
		err = mapstructure.Decode(sync.Schedule, &schedule)
		if err != nil {
			return err
		}
		resourceBlock.Body().SetAttributeValue("schedule", typeConverter(schedule))
		var fields []map[string]interface{}
		err = mapstructure.Decode(sync.Fields, &fields)
		if err != nil {
			return err
		}
		resourceBlock.Body().SetAttributeValue("fields", typeConverter(fields))
		var target map[string]interface{}
		err = mapstructure.Decode(sync.Target, &target)
		if err != nil {
			return err
		}
		tokens := wrapJSONEncode(target, "search_values", "configuration")
		resourceBlock.Body().SetAttributeRaw("target", tokens)

		if sync.FilterLogic != "" {
			resourceBlock.Body().SetAttributeValue("filter_logic", cty.StringVal(sync.FilterLogic))
		}

		if len(sync.Filters) > 0 {
			var filters []map[string]interface{}
			err = mapstructure.Decode(sync.Filters, &filters)
			if err != nil {
				return err
			}
			filterTokens := wrapJSONEncode(filters, "value")
			resourceBlock.Body().SetAttributeRaw("filters", filterTokens)
		}

		if sync.Identity != nil {
			var identity map[string]interface{}
			err = mapstructure.Decode(sync.Identity, &identity)
			if err != nil {
				return err
			}
			resourceBlock.Body().SetAttributeValue("identity", typeConverter(identity))
		}
		if len(sync.OverrideFields) > 0 {
			var overrideFields []map[string]interface{}
			err = mapstructure.Decode(sync.OverrideFields, &overrideFields)
			if err != nil {
				return err
			}
			resourceBlock.Body().SetAttributeValue("override_fields", typeConverter(overrideFields))
		}
		if len(sync.Overrides) > 0 {
			var overrides []map[string]interface{}
			err = mapstructure.Decode(sync.Overrides, &overrides)
			if err != nil {
				return err
			}
			resourceBlock.Body().SetAttributeValue("overrides", typeConverter(overrides))
		}
		resourceBlock.Body().SetAttributeValue("sync_all_records", cty.BoolVal(sync.SyncAllRecords))
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
			sync.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", sync.Name)))
	}
	return nil
}

func (s *Syncs) Filename() string {
	return SyncResourceFileName
}

func (s *Syncs) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, sync := range s.Resources {
		result[sync.ID] = fmt.Sprintf("%s.%s.id", SyncResource, name)
	}
	return result
}

func (s *Syncs) DatasourceRefs() map[string]string {
	return nil
}
