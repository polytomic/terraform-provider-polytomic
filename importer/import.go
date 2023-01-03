package importer

import (
	"bytes"
	"context"
	"text/template"

	"github.com/polytomic/polytomic-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	ImportFileName = "import.sh"
)

var (
	mainTemplate = `
terraform {
		required_providers {
		  polytomic = {
			source = "polytomic/polytomic"
		  }
		}
	  }

	  provider "polytomic" {
		deployment_url = "{{ .URL }}"
		deployment_api_key = "{{ .APIKey }}"
	  }
`
)

type Importable interface {
	Init(ctx context.Context) error
	GenerateTerraformFiles(ctx context.Context, file string, recreate bool) error
	GenerateImports(ctx context.Context, path string, recreate bool) (bytes.Buffer, error)
}

func Init(url string, key string, path string, recreate bool) {
	err := createDirectory(path)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to create directory")
	}

	ctx := context.Background()
	c := polytomic.NewClient(url, polytomic.APIKey(key))

	importables := map[string]Importable{
		"connections": NewConnections(c),
	}

	// Create import.sh
	importFile, err := createFile(recreate, 0755, path, ImportFileName)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to create import.sh")
	}
	defer importFile.Close()

	// Create main.tf
	err = createMainFile(path, recreate)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to create main.tf")
	}

	for _, i := range importables {
		err := i.Init(ctx)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to initialize")
		}
		err = i.GenerateTerraformFiles(ctx, path, recreate)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to generate terraform files")
		}
		imports, err := i.GenerateImports(ctx, path, recreate)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to generate imports")
		}
		importFile.Write(imports.Bytes())
	}

}

func createMainFile(path string, recreate bool) error {
	url := viper.GetString("url")
	apiKey := viper.GetString("api-key")

	f, err := createFile(recreate, 0644, path, "main.tf")
	if err != nil {
		return err
	}
	defer f.Close()
	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		URL    string
		APIKey string
	}{
		URL:    url,
		APIKey: apiKey,
	})
	if err != nil {
		return err
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil

}
