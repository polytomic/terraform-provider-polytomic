package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/polytomic/polytomic-go"
	"github.com/polytomic/terraform-provider-polytomic/provider"
	"github.com/zclconf/go-cty/cty"
)

const (
	ModelsResourceFileName = "models.tf"
	ModelResource          = "polytomic_model"
)

var (
	_ Importable = &Models{}
)

type Models struct {
	c *polytomic.Client
	// modelNames is a map of model id's to their disambiguated names
	modelNames  map[string]string
	uniqueNames map[string]bool

	Resources []*polytomic.Model
}

func NewModels(c *polytomic.Client) *Models {
	return &Models{
		c:           c,
		modelNames:  map[string]string{},
		uniqueNames: map[string]bool{},
	}
}

func (m *Models) Init(ctx context.Context) error {
	models, err := m.c.Models().List(ctx)
	if err != nil {
		return err
	}

	for _, model := range models {
		hydratedModel, err := m.c.Models().Get(ctx, model.ID)
		if err != nil {
			return err
		}

		m.Resources = append(m.Resources, hydratedModel)
	}

	return nil

}

func (m *Models) GenerateTerraformFiles(ctx context.Context, writer io.Writer) error {
	for _, model := range m.Resources {
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		name := provider.ValidNamer(
			provider.ToSnakeCase(model.Name),
		)
		if _, exists := m.uniqueNames[name]; exists {
			name = fmt.Sprintf("%s_%s", name, model.Type)
		}
		m.uniqueNames[name] = true
		m.modelNames[model.ID] = name

		resourceBlock := body.AppendNewBlock("resource", []string{ModelResource, name})
		resourceBlock.Body().SetAttributeValue("connection_id", cty.StringVal(model.ConnectionID))
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(model.Name))

		// Clean model configuration values before converting to cty types
		for k, v := range model.Configuration {
			// Remove empty values and tracking_columns
			if v == "" || k == "tracking_columns" {
				delete(model.Configuration, k)
			}
		}

		resourceBlock.Body().SetAttributeValue("configuration", typeConverter(model.Configuration))

		var modelFields []string
		var modelAdditionalFields []map[string]interface{}
		for _, field := range model.Fields {
			if !field.UserAdded {
				modelFields = append(modelFields, field.Name)
			} else {
				modelAdditionalFields = append(modelAdditionalFields, map[string]interface{}{
					"name":  field.Name,
					"type":  field.Type,
					"label": field.Label,
				})
			}
		}
		resourceBlock.Body().SetAttributeValue("fields", typeConverter(modelFields))
		resourceBlock.Body().SetAttributeValue("additional_fields", typeConverter(modelAdditionalFields))

		var modelRelations []map[string]interface{}
		for _, relation := range model.Relations {
			modelRelations = append(modelRelations, map[string]interface{}{
				"to": map[string]interface{}{
					"model_id": relation.RelationTo.ModelID,
					"field":    relation.RelationTo.Field,
				},
				"from": relation.From,
			})
		}

		resourceBlock.Body().SetAttributeValue("relations", typeConverter(modelRelations))
		resourceBlock.Body().SetAttributeValue("identifier", cty.StringVal(model.Identifier))
		resourceBlock.Body().SetAttributeValue("tracking_columns", typeConverter(model.TrackingColumns))

		writer.Write(hclFile.Bytes())
	}
	return nil
}

func (m *Models) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, model := range m.Resources {
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			ModelResource,
			m.modelNames[model.ID],
			model.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", model.Name)))
	}
	return nil
}

func (m *Models) Filename() string {
	return ModelsResourceFileName
}
