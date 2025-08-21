# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the Terraform provider for Polytomic, enabling infrastructure-as-code
management of Polytomic resources like connections, models, syncs, and policies.
The provider supports 100+ connection types through code generation.

## Key Commands

### Development

```bash
# Full development setup (installs provider, generates code, builds docs)
make dev

# Build and install provider locally
go install

# Generate connection code from templates
go generate

# Run unit tests
go test ./...

# Run acceptance tests (requires POLYTOMIC_API_KEY)
TF_ACC=1 go test ./... -v

# Generate documentation
go generate

# Format code
go fmt ./...
```

### Local Testing

1. Create a `.tfrc` file pointing to local provider:

```hcl
provider_installation {
  dev_overrides {
    "polytomic/polytomic" = "/Users/[username]/go/bin"
  }
  direct {}
}
```

2. Set environment: `export TF_CLI_CONFIG_FILE=.tfrc`
3. Provider will use version "99.0.0" for local development

## Architecture

### Directory Structure

- `/provider` - Core provider implementation
  - `/resource_*.go` - Resource implementations (connections, models, syncs, etc.)
  - `/data_source_*.go` - Data source implementations
  - `/gen/` - Code generation logic for connection types
  - `/internal/` - Generated connection implementations
  - `/client.go` - Polytomic API client wrapper
  - `/validators/` - Custom validators for provider configuration
- `/importer` - CLI tool for importing existing Polytomic configurations
  - `/main.go` - CLI entry point and provider template generation
  - `/import.go` - Core import orchestration and variable handling
  - `/connections.go` - Connection resource/datasource import logic
  - `/models.go`, `/syncs.go`, `/bulk_syncs.go` - Resource-specific importers
  - `/policy.go`, `/role.go` - Permission resource importers (optional)
  - `/variables.go` - Terraform variable generation utilities
  - `/formatter.go` - HCL formatting and variable reference handling
- `/hack` - Development scripts
- `/docs` - Auto-generated documentation
- `/examples` - Usage examples

### Code Generation Pattern

The provider uses extensive code generation for connection types:

1. Templates in `/provider/gen/internal/templates/`
2. Generator code in `/provider/gen/internal/generator/`
3. Connection definitions loaded from Polytomic API
4. Generated files in `/provider/internal/`
5. Run `go generate` to regenerate

### Authentication

Provider supports three authentication methods (in order of precedence):

1. Deployment key (for globa access; may be specified with an organization ID)
2. Partner key (for partner access; may be specified with an organization ID)
3. API key (standard organization access)

Configuration via environment variables or provider block:

```hcl
provider "polytomic" {
  deployment_key = "..."
  # OR
  partner_key = "..."
  account_id = "..."
  # OR
  api_key = "..."
}
```

### Resource Patterns

All resources follow standard Terraform CRUD patterns:

- Create: `resource<Type>Create()`
- Read: `resource<Type>Read()`
- Update: `resource<Type>Update()`
- Delete: `resource<Type>Delete()`

Resources use Terraform Plugin Framework with typed models.

### Testing Approach

- Unit tests: Standard Go tests in `*_test.go` files
- Acceptance tests: Real API tests with `TF_ACC=1`
- CI/CD: GitHub Actions on push and PR
- Local provider override via `.tfrc` for manual testing

## Common Tasks

### Adding a New Resource

1. Create `provider/resource_<name>.go`
2. Implement CRUD functions following existing patterns
3. Add to provider schema in `provider/provider.go`
4. Create acceptance test in `provider/resource_<name>_test.go`
5. Run `go generate` to update docs
6. Add example in `examples/resources/<name>/`

### Updating Connection Types

1. Connection definitions are fetched from Polytomic API
2. Modify generator code if needed in `/provider/gen/`
3. Run `go generate` to regenerate
4. Test with acceptance tests

### Working with the Importer

1. **Adding New Resource Types**:

   - Create new importer component implementing `Importable` interface
   - Add to `importables` slice in `import.go`
   - Handle organization tracking with `OrganizationTracker` if applicable

2. **Testing Importer Changes**:

   - Build importer: `go build -o importer ./importer`
   - Run with test API key: `./importer run --api-key $API_KEY --output test-import`
   - Check generated `.tf` files and `variables.tf`

3. **Variable Generation Patterns**:
   - Use centralized collection in `import.go` for shared variables
   - Implement `OrganizationTracker` interface for organization-aware components
   - Use `unquoteVariableRef()` for variable references in HCL output

### Debugging

- Enable debug logs: `TF_LOG=DEBUG terraform apply`
- API client logs requests/responses with debug logging
- Use `tflog` package for structured logging in provider code

## Release Process

1. Update CHANGELOG.md
2. Create and push git tag: `git tag v1.0.0 && git push origin v1.0.0`
3. GitHub Actions automatically builds and publishes to Terraform Registry

## Importer Architecture

### Component Pattern

The importer uses a modular architecture with the `Importable` interface:

```go
type Importable interface {
    Init(ctx context.Context) error
    ResourceRefs() map[string]string
    DatasourceRefs() map[string]string
    GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error
    GenerateImports(ctx context.Context, writer io.Writer) error
    Filename() string
    Variables() []Variable
}
```

### Organization Variable Pattern

The importer generates reusable variables for commonly referenced values:

- **Single Organization**: When all resources belong to one organization, creates `var.polytomic_organization_id`
- **Variable References**: Uses `unquoteVariableRef()` to convert quoted variable strings to actual variable references
- **Centralized Collection**: `OrganizationTracker` interface allows components to report organization IDs
- **Two-Pass Processing**: First pass collects metadata, second pass generates files with proper variable references

### HCL Generation

- Uses `hclwrite` package for programmatic HCL generation
- `typeConverter()` handles arbitrary value to `cty.Value` conversion
- `unquoteVariableRef()` post-processes HCL to convert `"var.name"` to `var.name`
- Variable references enable more maintainable generated code

### Authentication Context

- Single organization API key: Generates organization variable with default value
- Multi-organization scenarios: Falls back to hard-coded organization IDs
- Currently supports API key authentication only (not deployment keys)

## Important Notes

- Provider uses a fork of terraform-plugin-framework (see go.mod)
- All API operations use the Polytomic Go SDK
- Resource imports supported via `terraform import`
- Sensitive values (API keys, secrets) marked with `Sensitive: true`
- Connection configurations stored as JSON strings in state
- Importer generates variables for reusable values (organization IDs, etc.)
