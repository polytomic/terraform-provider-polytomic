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
	BulkSyncResourceFileName = "bulk_syncs.tf"
	BulkSyncResource         = "polytomic_bulk_sync"
)

var (
	_ Importable = &BulkSyncs{}
)

type BulkSyncs struct {
	c *polytomic.Client

	Resources map[string]polytomic.BulkSyncResponse
}

func NewBulkSyncs(c *polytomic.Client) *BulkSyncs {
	return &BulkSyncs{
		c:         c,
		Resources: make(map[string]polytomic.BulkSyncResponse),
	}
}

func (b *BulkSyncs) Init(ctx context.Context) error {
	bulkSyncs, err := b.c.Bulk().ListBulkSyncs(ctx)
	if err != nil {
		return err
	}
	for _, bulk := range bulkSyncs {
		// Bulk sync names are not unique, so we need to a slug to the name
		// to make it unique.
		name := provider.ValidName(provider.ToSnakeCase(bulk.Name) + "_" + bulk.ID[:8])
		b.Resources[name] = bulk
	}

	return nil
}

func (b *BulkSyncs) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	for _, name := range sortedKeys(b.Resources) {
		bulkSync := b.Resources[name]
		bulkSchemas, err := b.c.Bulk().GetBulkSyncSchemas(ctx, bulkSync.ID)
		if err != nil {
			return err
		}
		schemas := make([]string, 0, len(bulkSchemas))
		for _, schema := range bulkSchemas {
			if schema.Enabled {
				schemas = append(schemas, schema.ID)
			}
		}
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()

		resourceBlock := body.AppendNewBlock("resource", []string{BulkSyncResource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(bulkSync.Name))
		resourceBlock.Body().SetAttributeValue("source_connection_id", cty.StringVal(bulkSync.SourceConnectionID))
		resourceBlock.Body().SetAttributeValue("dest_connection_id", cty.StringVal(bulkSync.DestConnectionID))
		resourceBlock.Body().SetAttributeValue("active", cty.BoolVal(bulkSync.Active))
		resourceBlock.Body().SetAttributeValue("discover", cty.BoolVal(bulkSync.Discover))
		resourceBlock.Body().SetAttributeValue("mode", cty.StringVal(bulkSync.Mode))

		dTokens := wrapJSONEncode(bulkSync.DestinationConfiguration, "advanced")
		resourceBlock.Body().SetAttributeRaw("dest_configuration", dTokens)
		sTokens := wrapJSONEncode(bulkSync.SourceConfiguration, "advanced")
		resourceBlock.Body().SetAttributeRaw("source_configuration", sTokens)

		resourceBlock.Body().SetAttributeValue("schemas", typeConverter(schemas))

		var schedule map[string]*string
		err = mapstructure.Decode(bulkSync.Schedule, &schedule)
		if err != nil {
			return err
		}
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
			bulkSync.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", bulkSync.Name)))
	}
	return nil
}

func (b *BulkSyncs) Filename() string {
	return BulkSyncResourceFileName
}

func (b *BulkSyncs) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, bulk := range b.Resources {
		result[bulk.ID] = name
	}
	return result
}

func (b *BulkSyncs) DatasourceRefs() map[string]string {
	return nil
}
