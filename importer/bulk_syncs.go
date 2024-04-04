package importer

import (
	"context"
	"fmt"
	"io"

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
	bulkSyncs, err := b.c.BulkSync.List(ctx)
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
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()

		resourceBlock := body.AppendNewBlock("resource", []string{BulkSyncResource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(pointer.GetString(bulkSync.Name)))
		resourceBlock.Body().SetAttributeValue("source_connection_id", cty.StringVal(pointer.GetString(bulkSync.SourceConnectionId)))
		resourceBlock.Body().SetAttributeValue("dest_connection_id", cty.StringVal(pointer.GetString(bulkSync.DestinationConnectionId)))
		resourceBlock.Body().SetAttributeValue("active", cty.BoolVal(pointer.GetBool(bulkSync.Active)))
		resourceBlock.Body().SetAttributeValue("discover", cty.BoolVal(pointer.GetBool(bulkSync.Discover)))
		resourceBlock.Body().SetAttributeValue("mode", cty.StringVal(pointer.GetString(bulkSync.Mode)))

		dTokens := wrapJSONEncode(bulkSync.DestinationConfiguration, "advanced")
		resourceBlock.Body().SetAttributeRaw("dest_configuration", dTokens)
		sTokens := wrapJSONEncode(bulkSync.SourceConfiguration, "advanced")
		resourceBlock.Body().SetAttributeRaw("source_configuration", sTokens)

		resourceBlock.Body().SetAttributeValue("schemas", typeConverter(schemas))

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
