package importer

import (
	"context"
	"io"

	"github.com/polytomic/polytomic-go"
	"github.com/rs/zerolog/log"
)

const (
	ImportFileName = "import.sh"
)

type Importable interface {
	Init(ctx context.Context) error
	ResourceRefs() map[string]string
	DatasourceRefs() map[string]string
	GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error
	GenerateImports(ctx context.Context, writer io.Writer) error
	Filename() string
	Variables() []Variable
}

func Init(url, key, path string, recreate bool, includePermissions bool) {
	err := createDirectory(path)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to create directory")
	}

	ctx := context.Background()
	c := polytomic.NewClient(url, polytomic.APIKey(key))

	importables := []Importable{
		NewMain(),
		NewConnections(c),
		NewModels(c),
		NewBulkSyncs(c),
		NewSyncs(c),
	}

	if includePermissions {
		importables = append(importables, NewPolicies(c))
		importables = append(importables, NewRoles(c))
	}

	// Create import.sh
	importFile, err := createFile(recreate, 0755, path, ImportFileName)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to create import.sh")
	}
	defer importFile.Close()

	vars := []Variable{}
	refs := make(map[string]string)
	for _, i := range importables {
		err := i.Init(ctx)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to initialize")
		}
		// Add resource refs
		for k, v := range i.ResourceRefs() {
			refs[k] = v
		}
		// Add datasource refs
		for k, v := range i.DatasourceRefs() {
			refs[k] = v
		}
		// Add variables
		vars = append(vars, i.Variables()...)

		file := i.Filename()
		f, err := createFile(recreate, 0644, path, file)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to create file")
		}
		defer f.Close()

		err = i.GenerateTerraformFiles(ctx, f, refs)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to generate terraform files")
		}
		err = i.GenerateImports(ctx, importFile)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to generate imports")
		}
	}

	// Create variables.tf
	variablesFile, err := createFile(recreate, 0644, path, "variables.tf")
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to create variables.tf")
	}
	defer variablesFile.Close()

	err = generateVariables(variablesFile, vars)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to generate variables")
	}

}
