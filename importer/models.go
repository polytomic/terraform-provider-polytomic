package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
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
	c *ptclient.Client
	// modelNames is a map of model id's to their disambiguated names
	modelNames  map[string]string
	uniqueNames map[string]bool

	Resources map[string]*polytomic.ModelResponse
}

func NewModels(c *ptclient.Client) *Models {
	return &Models{
		c:           c,
		modelNames:  map[string]string{},
		uniqueNames: map[string]bool{},
		Resources:   make(map[string]*polytomic.ModelResponse),
	}
}

func (m *Models) Init(ctx context.Context) error {
	models, err := m.c.Models.List(ctx)
	if err != nil {
		return err
	}

	for _, model := range models.Data {
		hydratedModel, err := m.c.Models.Get(ctx, pointer.GetString(model.Id))
		if err != nil {
			return err
		}

		name := provider.ValidName(
			provider.ToSnakeCase(pointer.GetString(model.Name)),
		)
		if _, exists := m.uniqueNames[name]; exists {
			name = fmt.Sprintf("%s_%s", name, pointer.GetString(model.Type))
		}
		m.uniqueNames[name] = true
		m.modelNames[pointer.GetString(model.Id)] = name
		m.Resources[name] = hydratedModel.Data
	}

	return nil

}

func (m *Models) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	for _, name := range sortedKeys(m.Resources) {
		model := m.Resources[name]
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()

		resourceBlock := body.AppendNewBlock("resource", []string{ModelResource, name})
		resourceBlock.Body().SetAttributeValue("connection_id", cty.StringVal(pointer.GetString(model.ConnectionId)))
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(pointer.GetString(model.Name)))

		// Clean model configuration values before converting to cty types
		for k, v := range model.Configuration {
			// Remove empty values and tracking_columns
			if v == "" || k == "tracking_columns" {
				delete(model.Configuration, k)
			}
		}

		resourceBlock.Body().SetAttributeRaw("configuration", wrapJSONEncode(model.Configuration))

		var modelFields []string
		var modelAdditionalFields []map[string]interface{}
		for _, field := range model.Fields {
			if !pointer.GetBool(field.UserAdded) {
				modelFields = append(modelFields, pointer.GetString(field.Name))
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
					"model_id": relation.To.ModelId,
					"field":    relation.To.Field,
				},
				"from": relation.From,
			})
		}

		resourceBlock.Body().SetAttributeValue("relations", typeConverter(modelRelations))
		if pointer.GetString(model.Identifier) != "" {
			resourceBlock.Body().SetAttributeValue("identifier", cty.StringVal(pointer.GetString(model.Identifier)))
		}
		resourceBlock.Body().SetAttributeValue("tracking_columns", typeConverter(model.TrackingColumns))

		writer.Write(ReplaceRefs(hclFile.Bytes(), refs))
	}
	return nil
}

func (m *Models) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, name := range sortedKeys(m.Resources) {
		model := m.Resources[name]
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			ModelResource,
			m.modelNames[pointer.GetString(model.Id)],
			pointer.GetString(model.Id))))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", pointer.GetString(model.Name))))
	}
	return nil
}

func (m *Models) Filename() string {
	return ModelsResourceFileName
}

func (m *Models) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, model := range m.Resources {
		result[pointer.GetString(model.Id)] = fmt.Sprintf("%s.%s.id", ModelResource, name)
	}
	return result
}

func (m *Models) DatasourceRefs() map[string]string {
	return nil
}

func (m *Models) Variables() []Variable {
	return nil
}
