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

	Resources []polytomic.BulkSyncResponse
}

func NewBulkSyncs(c *polytomic.Client) *BulkSyncs {
	return &BulkSyncs{
		c: c,
	}
}

func (b *BulkSyncs) Init(ctx context.Context) error {
	bulkSyncs, err := b.c.Bulk().ListBulkSyncs(ctx)
	if err != nil {
		return err
	}
	b.Resources = bulkSyncs

	return nil
}

func (b *BulkSyncs) GenerateTerraformFiles(ctx context.Context, writer io.Writer) error {
	for _, bulkSync := range b.Resources {
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
		// Bulk sync names are not unique, so we need to a slug to the name
		// to make it unique.
		name := provider.ValidName(provider.ToSnakeCase(bulkSync.Name) + "_" + bulkSync.ID[:8])
		resourceBlock := body.AppendNewBlock("resource", []string{BulkSyncResource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(bulkSync.Name))
		resourceBlock.Body().SetAttributeValue("source_connection_id", cty.StringVal(bulkSync.SourceConnectionID))
		resourceBlock.Body().SetAttributeValue("dest_connection_id", cty.StringVal(bulkSync.DestConnectionID))
		resourceBlock.Body().SetAttributeValue("active", cty.BoolVal(bulkSync.Active))
		resourceBlock.Body().SetAttributeValue("discover", cty.BoolVal(bulkSync.Discover))
		resourceBlock.Body().SetAttributeValue("mode", cty.StringVal(bulkSync.Mode))
		resourceBlock.Body().SetAttributeValue("dest_configuration", typeConverter(bulkSync.DestinationConfiguration))
		resourceBlock.Body().SetAttributeValue("source_configuration", typeConverter(bulkSync.SourceConfiguration))
		resourceBlock.Body().SetAttributeValue("schemas", typeConverter(schemas))

		var schedule map[string]*string
		err = mapstructure.Decode(bulkSync.Schedule, &schedule)
		if err != nil {
			return err
		}
		resourceBlock.Body().SetAttributeValue("schedule", typeConverter(schedule))
		body.AppendNewline()

		writer.Write(hclFile.Bytes())
	}

	return nil
}

func (b *BulkSyncs) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, bulkSync := range b.Resources {
		// Bulk sync names are not unique, so we need to a slug to the name
		// to make it unique.
		name := provider.ValidName(provider.ToSnakeCase(bulkSync.Name) + "_" + bulkSync.ID[:8])
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
