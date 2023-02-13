package importer

import (
	"context"
	"fmt"
	"io"
	"strings"

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

	Resources []polytomic.SyncResponse
}

func NewSyncs(c *polytomic.Client) *Syncs {
	return &Syncs{
		c: c,
	}
}

func (s *Syncs) Init(ctx context.Context) error {
	syncs, err := s.c.Syncs().List(ctx)
	if err != nil {
		return err
	}

	s.Resources = append(s.Resources, syncs...)

	return nil
}

func (s *Syncs) GenerateTerraformFiles(ctx context.Context, writer io.Writer) error {
	for _, sync := range s.Resources {
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{SyncResource, provider.ToSnakeCase(sync.Name)})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(sync.Name))
		resourceBlock.Body().SetAttributeValue("mode", cty.StringVal(sync.Mode))
		var schedule map[string]*string
		err := mapstructure.Decode(sync.Schedule, &schedule)
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
		resourceBlock.Body().SetAttributeValue("target", typeConverter(target))
		if sync.FilterLogic != "" {
			resourceBlock.Body().SetAttributeValue("filter_logic", cty.StringVal(sync.FilterLogic))
		}
		if len(sync.Filters) > 0 {
			var res string
			res += "["
			for _, filter := range sync.Filters {
				if filter.Value != nil {
					tmpl := `{
					field_id = "%s"
					field_type = "%s"
					function = "%s"
					value = jsonencode(%s)
				},`

					var v string
					switch filter.Value.(type) {
					case []interface{}:
						var val []string
						for _, v := range filter.Value.([]interface{}) {
							val = append(val, v.(string))
						}
						v = fmt.Sprintf(`["%s"]`, strings.Join(val, `","`))
					default:
						v = fmt.Sprintf(`"%s"`, filter.Value.(string))
					}
					res += fmt.Sprintf(tmpl, filter.FieldID, filter.FieldType, filter.Function, v)

				} else {
					tmpl := `{
					field_id = "%s"
					field_type = "%s"
					function = "%s"
				},`
					res += fmt.Sprintf(tmpl, filter.FieldID, filter.FieldType, filter.Function)
				}
			}
			res += "]"

			resourceBlock.Body().SetAttributeRaw("filters", hclwrite.Tokens{{
				Bytes: []byte(res),
			}})
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

		writer.Write(hclFile.Bytes())

	}
	return nil
}

func (s *Syncs) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, sync := range s.Resources {
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			SyncResource,
			provider.ToSnakeCase(sync.Name),
			sync.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", sync.Name)))
	}
	return nil
}

func (s *Syncs) Filename() string {
	return SyncResourceFileName
}
