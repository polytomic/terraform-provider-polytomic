package importer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/polytomic/polytomic-go"
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
	# Configuration comes from environment variables:
	# POLYTOMIC_DEPLOYMENT_URL
	# POLYTOMIC_API_KEY or POLYTOMIC_DEPLOYMENT_KEY or POLYTOMIC_PARTNER_KEY
}

{{ if .OrgResource }}
	resource "polytomic_organization" "{{ .Slug }}" {
		name = "{{ .Name }}"
	}

	locals {
		organization_id = polytomic_organization.{{ .Slug }}.id
	}
{{ else }}
    data "polytomic_caller_identity" "self" {}

	locals {
		organization_id = data.polytomic_caller_identity.self.organization_id
	}
{{ end -}}
`
)

func NewMain(org *polytomic.Organization, orgResource bool) *Main {
	slug := strings.TrimSpace(pointer.Get(org.Name))
	if slug == "" {
		panic("organization name is empty")
	}

	m := &Main{
		OrgResource: orgResource,
		Slug:        pointer.Get(org.Name),
		Name:        pointer.Get(org.Name),
	}
	if orgResource && pointer.Get(org.Id) != "" {
		m.ID = pointer.Get(org.Id)
	}

	return m
}

type Main struct {
	OrgResource bool
	ID          string
	Slug        string
	Name        string
}

func (m *Main) Init(ctx context.Context) error {
	return nil
}

func (m *Main) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, m)
	if err != nil {
		return err
	}
	_, err = writer.Write(hclwrite.Format(buf.Bytes()))
	return err
}

func (m *Main) GenerateImports(ctx context.Context, writer io.Writer) error {
	if m.ID == "" {
		return nil
	}
	_, err := fmt.Fprintf(writer, "terraform import polytomic_organization.%s %s # %s\n", m.Slug, m.ID, m.Name)
	return err
}

func (m *Main) Filename() string {
	return "main.tf"
}

func (m *Main) ResourceRefs() map[string]string {
	result := map[string]string{}
	if m.ID != "" {
		result[m.ID] = fmt.Sprintf("polytomic_organization.%s.id", m.Slug)
	}
	return result
}

func (m *Main) DatasourceRefs() map[string]string {
	return nil
}

func (m *Main) Variables() []Variable {
	return nil
}
