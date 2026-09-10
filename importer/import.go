package importer

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/polytomic/polytomic-go/v25"
	ptclient "github.com/polytomic/polytomic-go/v25/client"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
	"github.com/rs/zerolog/log"
)

const (
	UserAgent      = "polytomic-terraform-provider/importer"
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

func Init(ctx context.Context, clientProvider *providerclient.Provider, organizations, outputPath string, recreate, includePermissions, includeSchemaOverrides bool) {
	err := createDirectory(outputPath)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to create directory")
	}

	// // Handle organization discovery and filtering
	orgFilter := make(map[string]bool)
	for _, id := range strings.Split(organizations, ",") {
		if strings.TrimSpace(id) != "" {
			orgFilter[strings.TrimSpace(id)] = true
		}
	}

	// Discover all accessible organizations
	orgs, err := clientProvider.ListOrganizations(ctx)
	if err != nil {
		log.Fatal().AnErr("error", err).Msg("failed to list organizations")
	}
	targetOrgs := make([]*polytomic.Organization, 0, len(orgs))

	// Filter organizations if specified, otherwise use all discovered
	// organizations
	for _, org := range orgs {
		if len(orgFilter) == 0 || orgFilter[pointer.Get(org.ID)] {
			targetOrgs = append(targetOrgs, org)
		}
	}

	if len(targetOrgs) == 0 {
		log.Fatal().Msg("no matching organizations found")
	}
	if len(orgFilter) == 0 {
		// Log discovered organizations for user awareness
		log.Info().Msgf("Discovered %d organization(s) for export:", len(targetOrgs))
		for _, org := range targetOrgs {
			log.Info().Str("id", pointer.Get(org.ID)).Str("name", pointer.Get(org.Name)).Msg("  - Organization")
		}
	}

	// Handle single vs multi-org mode
	if len(targetOrgs) > 1 {
		// Multi-org mode: create separate directories
		for _, org := range targetOrgs {
			orgPath := filepath.Join(outputPath, pointer.Get(org.Name))

			// Import resources for this organization
			orgClient, err := clientProvider.Client(ctx, pointer.Get(org.ID))
			if err != nil {
				log.Fatal().AnErr("error", err).Msg("failed to create organization client")
			}
			importOrganization(ctx, org, orgClient, orgPath, recreate, includePermissions, includeSchemaOverrides, true)
		}
	} else {
		// Single organization - use it directly
		orgClient, err := clientProvider.Client(ctx, pointer.Get(targetOrgs[0].ID))
		if err != nil {
			log.Fatal().AnErr("error", err).Msg("failed to create organization client")
		}
		importOrganization(ctx, targetOrgs[0], orgClient, outputPath, recreate, includePermissions, includeSchemaOverrides, false)
	}
}

// importOrganization imports resources for a single organization
func importOrganization(ctx context.Context, org *polytomic.Organization, c *ptclient.Client, path string, recreate, includePermissions, includeSchemaOverrides, orgResource bool) {
	log.Info().
		Str("org_id", pointer.Get(org.ID)).
		Str("org_name", pointer.Get(org.Name)).
		Str("path", path).
		Msg("importing organization")
	err := createDirectory(path)
	if err != nil {
		log.Error().AnErr("error", err).Msg("failed to create directory")
		return
	}

	importables := []Importable{
		NewMain(org, orgResource),
		NewGlobalErrorSubscribers(c),
		NewConnections(c),
		NewModels(c),
		NewBulkSyncs(c),
		NewSyncs(c),
	}

	if includePermissions {
		importables = append(importables, NewRoles(c))
		importables = append(importables, NewPolicies(c))
	}

	if includeSchemaOverrides {
		importables = append(importables, NewSchemaOverrides(c, pointer.Get(org.ID)))
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
		log.Info().Str("filename", i.Filename()).Msg("importing")
		err := i.Init(ctx)
		if err != nil {
			log.Fatal().AnErr("error", err).
				Str("organization_id", *org.ID).
				Str("importable", i.Filename()).
				Msg("failed to initialize")
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
