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

	for _, sync := range syncs {
		s.Resources = append(s.Resources, sync)
	}

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
			var filters []map[string]interface{}
			err = mapstructure.Decode(sync.Filters, &filters)
			if err != nil {
				return err
			}
			resourceBlock.Body().SetAttributeValue("filters", typeConverter(filters))
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
