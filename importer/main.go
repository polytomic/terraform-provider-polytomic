package importer

import (
	"bytes"
	"context"
	"io"
	"text/template"

	"github.com/spf13/viper"
)

var (
	_ Importable = &Main{}

	mainTemplate = `terraform {
	required_providers {
		polytomic = {
			source = "polytomic/polytomic"
		}
	}
}

provider "polytomic" {
	deployment_url = "{{ .URL }}"
	{{- if .WriteAPIKey }}
	api_key = "{{ .APIKey }}"
	{{- else }}
	api_key = var.polytomic_api_key
	{{- end }}
}

{{- if not .WriteAPIKey }}
variable "polytomic_api_key" {
	type = string
}
{{- end }}
`
)

func NewMain() *Main {
	return &Main{}
}

type Main struct {
	URL         string
	APIKey      string
	WriteAPIKey bool
}

func (m *Main) Init(ctx context.Context) error {
	m.URL = viper.GetString("url")
	m.APIKey = viper.GetString("api-key")
	m.WriteAPIKey = viper.GetBool("with-api-key")

	return nil
}

func (m *Main) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		URL         string
		APIKey      string
		WriteAPIKey bool
	}{
		URL:         m.URL,
		APIKey:      m.APIKey,
		WriteAPIKey: m.WriteAPIKey,
	})
	if err != nil {
		return err
	}
	_, err = writer.Write(buf.Bytes())
	return err
}

func (m *Main) GenerateImports(ctx context.Context, writer io.Writer) error {
	return nil
}

func (m *Main) Filename() string {
	return "main.tf"
}

func (m *Main) ResourceRefs() map[string]string {
	return nil
}

func (m *Main) DatasourceRefs() map[string]string {
	return nil
}
