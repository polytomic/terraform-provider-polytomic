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
	GenerateTerraformFiles(ctx context.Context, writer io.Writer) error
	GenerateImports(ctx context.Context, writer io.Writer) error
	Filename() string
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

	for _, i := range importables {
		err := i.Init(ctx)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to initialize")
		}
		file := i.Filename()
		f, err := createFile(recreate, 0644, path, file)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to create file")
		}
		defer f.Close()

		err = i.GenerateTerraformFiles(ctx, f)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to generate terraform files")
		}
		err = i.GenerateImports(ctx, importFile)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to generate imports")
		}
	}

}
