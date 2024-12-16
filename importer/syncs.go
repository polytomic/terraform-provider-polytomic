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
	for _, name := range sortedKeys(s.Resources) {
		syn := s.Resources[name]
		sync, err := s.c.ModelSync.Get(ctx, pointer.GetString(syn.Id))
		if err != nil {
			return err
		}
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{SyncResource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(pointer.GetString(sync.Data.Name)))
		resourceBlock.Body().SetAttributeValue("active", cty.BoolVal(pointer.GetBool(sync.Data.Active)))
		resourceBlock.Body().SetAttributeValue("mode", cty.StringVal(pointer.GetString(sync.Data.Mode)))
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
