package importer

import (
	"context"
	"os"

	"github.com/polytomic/polytomic-go"
	"github.com/rs/zerolog/log"
)

type Importable interface {
	Init(ctx context.Context) error
	GenerateTerraformFiles(ctx context.Context, file string, recreate bool) error
	Imports(ctx context.Context, path string, recreate bool) error
}

func Init(url string, key string, path string, recreate bool) {
	ctx := context.Background()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info().Msgf("Creating directory %s", path)
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Fatal().AnErr("error", err).Msgf("failed to create directory %s", path)
		}
	}

	c := polytomic.NewClient(url, polytomic.APIKey(key))

	importables := map[string]Importable{
		"connections": NewConnections(c),
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
		err = i.Imports(ctx, path, recreate)
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to generate imports")
		}
	}
}
