# Terraform Provider Round-Trip Testing Framework

This directory contains a comprehensive testing framework to validate that the
Polytomic Terraform importer generates configurations that can be successfully
re-imported without drift. The framework ensures complete round-trip
compatibility between resources created via Terraform and those exported using
the importer.

## Overview

The round-trip testing framework validates the following workflow:

1. **Create** resources using Terraform provider
2. **Export** the resources using the Polytomic importer
3. **Import** the exported configuration in a clean workspace
4. **Validate** that no drift is detected between original and imported state

## Architecture

### Key Components

- **`types.go`** - Core types and configuration structs
- **`roundtrip.go`** - Main round-trip validation logic and utilities
- **`*_test.go`** - Specific test cases for different resource types
- **`Makefile`** - Convenient test execution commands

### Integration with Terraform Plugin Testing

The framework builds upon Terraform's `terraform-plugin-testing` library:

- Reuses existing provider test infrastructure
- Leverages standard TestCase/TestStep patterns
- Integrates with ImportState testing capabilities
- Shares test fixtures and validation logic

## Getting Started

### Prerequisites

1. **Environment Variables**:

   ```bash
   export TF_ACC=1
   export POLYTOMIC_API_KEY="your-api-key"
   export POLYTOMIC_DEPLOYMENT_URL="https://app.polytomic.com"
   ```

2. **Dependencies**:
   - Go 1.21+
   - Terraform CLI
   - Access to Polytomic API

### Running Tests

```bash
# Run all round-trip tests
make test

# Run with debugging
make test-debug

# Quick validation
make test-quick
```

### Basic Usage Example

```go
func TestAccRoundTrip_YourResource(t *testing.T) {
    fixtures := &TestFixtures{}

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccRoundTripPreCheck(t) },
        ProtoV6ProviderFactories: provider.GetTestAccProtoV6ProviderFactories(),
        Steps: []resource.TestStep{
            // Step 1: Create resources
            {
                Config: fixtures.YourResourceConfig("test"),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("polytomic_your_resource.test", "name", "test"),
                ),
            },
            // Step 2: Validate round-trip
            {
                Config: fixtures.YourResourceConfig("test"),
                Check: testAccRoundTripValidation(
                    []string{"polytomic_your_resource.test"},
                    fixtures.DefaultRoundTripOptions(),
                ),
            },
        },
    })
}
```

## Test Configuration

### RoundTripOptions

Configure validation behavior:

```go
type RoundTripOptions struct {
    IncludePermissions bool     // Include RBAC resources in export
    ValidateSensitive  bool     // Validate sensitive fields as variables
    IgnoreFields       []string // Fields to skip during comparison
    ExpectedVariables  []string // Variables that should be present
}
```

### Field Handling Strategy

| Field Type                                | Behavior             | Validation                  |
| ----------------------------------------- | -------------------- | --------------------------- |
| **Server-generated** (`id`, `created_at`) | Skip comparison      | Ignored                     |
| **Sensitive** (`password`, `api_key`)     | Export as variables  | Verify variable references  |
| **References** (`connection_id`)          | Terraform references | Validate reference syntax   |
| **Computed** (`status`, `fields`)         | Read-only            | Skip or verify read-only    |
| **JSON/Objects** (`configuration`)        | Deep comparison      | Normalize before comparison |

## Available Test Fixtures

The `TestFixtures` struct provides reusable configurations:

```go
fixtures := &TestFixtures{}

// Basic resources
fixtures.BasicConnection("name")
fixtures.ConnectionWithModel("prefix")
fixtures.FullSyncChain("prefix")

// Complex scenarios
fixtures.MultipleConnections("prefix", 3)
fixtures.ConnectionWithSensitiveConfig("name")
fixtures.BulkSync("prefix")
fixtures.WithSchedule("prefix")

// Configuration helpers
fixtures.DefaultRoundTripOptions()
fixtures.SensitiveRoundTripOptions()
fixtures.PermissionsRoundTripOptions()
```

## Test Scenarios

### 1. Basic Resource Tests

- Single connection import/export
- Model → Connection references
- Sync → Model → Connection chains

### 2. Complex Scenarios

- Multiple independent resources
- Cross-resource dependencies
- Bulk operations (10+ resources)

### 3. Edge Cases

- Sensitive field handling
- Special characters in names
- Large configuration objects
- Schedule and complex nested objects

### 4. Validation Tests

- Resource reference integrity
- Variable generation for sensitive fields
- Terraform syntax validity
- No-drift guarantee

## Validation Process

### 1. Export Phase

```bash
# Importer generates:
main.tf          # Resource definitions
variables.tf     # Variable declarations
terraform.tfvars # Variable values (if --with-api-key)
import.sh        # Import commands
```

### 2. Import Phase

```bash
# Clean workspace setup:
terraform init
bash import.sh   # Import all resources
terraform plan   # Verify no changes
```

### 3. Validation Checks

- **Syntax validation**: `terraform fmt -check`
- **Reference validation**: All resource references exist
- **Variable validation**: Sensitive fields use variables
- **Field comparison**: Original vs imported state
- **Drift detection**: `terraform plan` shows no changes

## Error Handling

### Expected Failures

- **API Rate Limits**: Automatic retry with exponential backoff
- **Transient Network Issues**: Retry logic with timeout
- **Resource Conflicts**: Cleanup and retry
- **Import Errors**: Detailed logging and state recovery

### Debugging

```bash
# Keep workspaces for inspection
TEST_KEEP_WORKSPACES=true make test-debug

# Maximum logging
TF_LOG=TRACE TF_LOG_PROVIDER=TRACE make test-debug

# Run specific test with debugging
make test-run TEST=TestAccRoundTrip_Connection
```

Debug artifacts are saved to:

- Export directory: `/tmp/roundtrip-export-*`
- Import workspace: `/tmp/roundtrip-import-*`

## Performance Targets

- **Basic scenario**: < 30 seconds
- **Complex scenario**: < 2 minutes
- **Full test suite**: < 15 minutes

## CI/CD Integration

### GitHub Actions

The framework includes CI workflows:

```yaml
# .github/workflows/roundtrip-tests.yml
- Basic round-trip validation on PRs
- Full test matrix on main branch
- Coverage reporting
- Parallel execution for performance
```

### Required Secrets

- `POLYTOMIC_API_KEY`
- `POLYTOMIC_DEPLOYMENT_URL`

## Extending the Framework

### Adding New Test Cases

1. **Create test file**: `roundtrip_newresource_test.go`
2. **Add fixture**: Method in `fixtures.go`
3. **Configure validation**: Set appropriate `RoundTripOptions`
4. **Update Makefile**: Add new test target

### Custom Validation Logic

```go
// Custom field comparison
func validateCustomField(original, imported interface{}) error {
    // Your validation logic
    return nil
}

// Add to validation results
results = append(results, ValidationResult{
    Resource: "polytomic_custom_resource",
    Field:    "custom_field",
    Valid:    validateCustomField(origValue, impValue) == nil,
})
```

### Resource-Specific Handling

```go
// Add to validateResourceFields()
switch original.Type {
case "polytomic_custom_resource":
    return validateCustomResource(original, imported, opts)
default:
    return validateGenericResource(original, imported, opts)
}
```

## Success Metrics

### Required Outcomes

- ✅ 100% of basic scenarios pass without drift
- ✅ 95% of advanced scenarios pass (documented exceptions)
- ✅ No sensitive data exposed in generated configs
- ✅ All resource references maintain integrity
- ✅ Generated code passes `terraform fmt/validate`
- ✅ Import scripts execute without errors

### Quality Gates

- All tests pass in CI
- No false positives in drift detection
- Sensitive fields properly handled
- Performance targets met

## Troubleshooting

### Common Issues

1. **Import cycle errors**: Ensure test helpers are in non-test files
2. **Provider factory not found**: Check `provider/testing.go` exports
3. **API timeouts**: Reduce parallel execution or add retries
4. **Drift detected**: Check field ignore lists and computed fields
5. **Variables not found**: Verify `ExpectedVariables` configuration

### Getting Help

1. Check test output logs
2. Inspect saved workspaces (with `TEST_KEEP_WORKSPACES=true`)
3. Review validation results for specific field failures
4. Run tests individually to isolate issues

## Contributing

When adding new tests:

1. Follow existing patterns and naming conventions
2. Add appropriate documentation
3. Include both positive and negative test cases
4. Ensure tests are deterministic and reliable
5. Update this README for new features

## Future Enhancements

- **Extended ImportState Testing**: More ImportState features
- **State Migration Testing**: Provider version upgrades
- **Performance Benchmarking**: Track import/export times
- **Fuzzing Integration**: Random configuration generation
- **Multi-Provider Testing**: Cross-provider resource dependencies
