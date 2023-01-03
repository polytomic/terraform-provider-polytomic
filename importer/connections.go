package importer

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/rs/zerolog/log"

	"github.com/polytomic/polytomic-go"
	"github.com/polytomic/terraform-provider-polytomic/provider"
	"github.com/polytomic/terraform-provider-polytomic/provider/gen/connections"
	"gopkg.in/yaml.v2"
)

const (
	ConnectionsResourceFileName = "connections.tf"
)

var (
	_ Importable = &Connections{}

	connTemplate = `
resource "{{ .Resource }}" "{{ .ResourceName }}" {
	  name = "{{ .Name }}"
	  organization = "{{ .Organization }}"
	  configuration = {
		{{- range .Config }}
		{{- if eq .Type "string" }}
		{{ .Name }} = "{{ .Value }}"
		{{- else }}
		{{ .Name }} = {{ .Value }}
		{{- end }}
		{{- end }}
	  }
}`
)

type Connections struct {
	c *polytomic.Client

	Resources []Connection
}

type Connection struct {
	ID           string
	Resource     string
	ResourceName string
	Name         string
	Organization string
	Schema       connections.Connection
	SourceData   polytomic.Connection
	Config       []Config
}

type Config struct {
	Name  string
	Type  string
	Value interface{}
}

func NewConnections(c *polytomic.Client) *Connections {
	return &Connections{
		c: c,
	}
}

func (c *Connections) Init(ctx context.Context) error {
	// Polytomic connections
	conns, err := c.c.Connections().List(ctx)
	if err != nil {
		return err
	}

	// Read connection schema
	config, err := ioutil.ReadFile(connections.ConnectionsFile)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to find connections file")
	}
	data := connections.Connections{}
	err = yaml.Unmarshal(config, &data)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to read connections file")
	}
	connectionLookup := make(map[string]connections.Connection)
	for _, r := range data.Connections {
		for i, a := range r.Attributes {
			t, ok := connections.TypeMap[a.Type]
			if !ok {
				log.Warn().Str("type", a.Type).Msg("unknown type, skipping")
				continue
			}
			r.Attributes[i].TfType = t.TfType
			r.Attributes[i].AttrType = t.AttrType
			r.Attributes[i].AttrName = provider.ToSnakeCase(a.Name)
		}
		connectionLookup[r.Connection] = r
	}

	for _, conn := range conns {
		// Get the connection from the lookup
		connectionSchema, ok := connectionLookup[conn.Type.ID]
		if !ok {
			log.Warn().Str("connection", conn.Type.ID).Msg("unknown connection, skipping")
			continue
		}
		if !connectionSchema.Resource {
			log.Warn().Str("connection", conn.Type.ID).Msg("connection is not a resource, skipping")
			continue
		}
		// Construct a hydrated connection config
		config := []Config{}
		for _, a := range connectionSchema.Attributes {
			// Get the value from the connection
			v, ok := conn.Configuration.(map[string]interface{})[a.AttrName]
			if !ok {
				// See if there is an alias for the attribute
				v, ok = conn.Configuration.(map[string]interface{})[a.Alias]
				if !ok {
					log.Warn().Str("connection", conn.Type.ID).Str("attribute", a.AttrName).Msg("attribute not found in connection, skipping")
					continue
				}
			}
			if a.Sensitive {
				v = "SENSITIVE"
			}
			config = append(config, Config{
				Name:  a.AttrName,
				Value: v,
				Type:  a.Type,
			})
		}

		c.Resources = append(c.Resources, Connection{
			ID:           conn.ID,
			Resource:     connectionSchema.Connection,
			ResourceName: provider.ToSnakeCase(conn.Name),
			Name:         conn.Name,
			Organization: conn.OrganizationId,
			Config:       config,
			Schema:       connectionSchema,
		})
	}

	return nil
}

func (c *Connections) GenerateTerraformFiles(ctx context.Context, path string, recreate bool) error {
	f, err := createFile(recreate, 0644, path, ConnectionsResourceFileName)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, conn := range c.Resources {
		tmpl, err := template.New("template").Parse(connTemplate)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, Connection{
			Resource:     connections.TerraformResourceName(conn.Schema.Connection),
			ResourceName: provider.ToSnakeCase(conn.Name),
			Name:         conn.Name,
			Organization: conn.Organization,
			Config:       conn.Config,
		})
		if err != nil {
			return err
		}
		_, err = f.Write(buf.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Connections) GenerateImports(ctx context.Context, path string, recreate bool) (bytes.Buffer, error) {
	var buf bytes.Buffer
	for _, conn := range c.Resources {
		buf.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			connections.TerraformResourceName(conn.Schema.Connection),
			conn.ResourceName,
			conn.ID)))
		buf.Write([]byte(fmt.Sprintf(" # %s\n", conn.Name)))
	}
	return buf, nil
}
