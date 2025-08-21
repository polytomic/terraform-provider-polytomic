# Terraform Provider-Importer Round-Trip Testing Framework

## Overview

This document outlines a comprehensive testing framework to validate that the Polytomic Terraform importer generates configurations that can be successfully re-imported without drift. The framework leverages Terraform's built-in acceptance testing framework (`terraform-plugin-testing`) to ensure complete round-trip compatibility between resources created via Terraform and those exported using the importer.

## Goals

1. **Validate Correctness**: Ensure imported configurations accurately represent the original resources
2. **Detect Drift**: Identify any differences between original and re-imported resources
3. **Handle Edge Cases**: Test sensitive fields, complex dependencies, and special configurations
4. **Ensure Maintainability**: Generate idiomatic, maintainable Terraform code
5. **Support CI/CD**: Enable automated testing in continuous integration pipelines
6. **Reuse Test Infrastructure**: Leverage existing acceptance test framework and fixtures

## Architecture

### Integration with Terraform Acceptance Testing

The framework builds upon Terraform's `terraform-plugin-testing` library, which is already used by the provider for acceptance tests. This approach allows us to:

- **Reuse existing test fixtures** from provider acceptance tests
- **Leverage built-in ImportState testing** capabilities
- **Use standard TestCase/TestStep patterns** familiar to Terraform developers
- **Share test infrastructure** (PreCheck, ProviderFactories, etc.)

### Core Components

#### 1. Enhanced Test Step (`provider/roundtrip_test.go`)

Extends the standard acceptance test framework with custom TestSteps for round-trip validation:

```go
// Custom TestStep that runs after resource creation
type ImporterTestStep struct {
    // Run the importer and validate generated files
    ExportPath string
    ValidateExport func(t *testing.T, exportPath string) error
    
    // Re-import and validate no drift
    ImportWorkspace string
    ValidateImport func(t *testing.T, state *terraform.State) error
}

// Integration with standard TestCase
func TestAccRoundTrip_Connection(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create resource
            {
                Config: testAccConnectionConfig("test"),
                Check: resource.ComposeTestCheckFunc(
                    testAccConnectionExists("polytomic_csv_connection.test"),
                ),
            },
            // Step 2: Test standard Terraform import
            {
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateVerifyIgnore: []string{"configuration.password"},
                ResourceName:      "polytomic_csv_connection.test",
            },
            // Step 3: Custom round-trip validation
            {
                Config: testAccConnectionConfig("test"),
                Check: testAccRoundTripValidation(
                    "polytomic_csv_connection.test",
                    RoundTripOptions{
                        IncludePermissions: false,
                        ValidateSensitive:  true,
                    },
                ),
            },
        },
    })
}
```

#### 2. Shared Test Configurations

Leverage existing test configurations and create reusable fixtures:

```go
// provider/test_fixtures.go
package provider

// Reusable test configurations
func testAccConnectionConfig(name string) string {
    return fmt.Sprintf(`
resource "polytomic_csv_connection" "test" {
    name = "%s"
    configuration = {
        url = "https://example.com/data.csv"
    }
}`, name)
}

func testAccComplexConfig(prefix string) string {
    return fmt.Sprintf(`
%s

resource "polytomic_model" "test" {
    name = "%s_model"
    connection_id = polytomic_csv_connection.test.id
    configuration = jsonencode({
        table = "users"
    })
}

resource "polytomic_sync" "test" {
    name = "%s_sync"
    model_id = polytomic_model.test.id
    target = {
        connection_id = polytomic_csv_connection.test.id
    }
}`, testAccConnectionConfig(prefix), prefix, prefix)
}
```

#### 3. Round-Trip Validation Functions (`provider/roundtrip_utils.go`)

Helper functions that integrate with the acceptance test framework:

```go
package provider

import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/terraform"
    "github.com/polytomic/terraform-provider-polytomic/importer"
)

type RoundTripOptions struct {
    IncludePermissions bool
    ValidateSensitive  bool
    IgnoreFields       []string
}

// testAccRoundTripValidation runs the importer and validates round-trip
func testAccRoundTripValidation(resourceName string, opts RoundTripOptions) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        // Step 1: Get current resource from state
        rs, ok := s.RootModule().Resources[resourceName]
        if !ok {
            return fmt.Errorf("resource not found: %s", resourceName)
        }
        
        // Step 2: Run importer to generate Terraform files
        exportDir := t.TempDir()
        err := runImporter(exportDir, opts.IncludePermissions)
        if err != nil {
            return fmt.Errorf("importer failed: %w", err)
        }
        
        // Step 3: Create new workspace and import
        importWS := setupTempWorkspace(t)
        defer cleanupWorkspace(importWS)
        
        // Copy generated files
        copyGeneratedFiles(exportDir, importWS.Dir)
        
        // Run import script
        err = executeImportScript(importWS, filepath.Join(exportDir, "import.sh"))
        if err != nil {
            return fmt.Errorf("import script failed: %w", err)
        }
        
        // Step 4: Validate no drift
        plan, err := terraformPlan(importWS)
        if err != nil {
            return fmt.Errorf("plan failed: %w", err)
        }
        
        if !plan.Empty() {
            return fmt.Errorf("unexpected drift detected: %s", plan.Summary())
        }
        
        // Step 5: Validate specific fields
        importedState := getWorkspaceState(importWS)
        return validateResourceFields(rs, importedState, opts)
    }
}

// runImporter executes the importer CLI
func runImporter(outputDir string, includePermissions bool) error {
    // Get credentials from environment (same as provider tests)
    apiKey := os.Getenv("POLYTOMIC_API_KEY")
    url := os.Getenv("POLYTOMIC_DEPLOYMENT_URL")
    
    // Initialize and run importer
    importer.Init(url, apiKey, outputDir, true, includePermissions)
    return nil
}
```

## Implementation Plan

### Phase 1: Foundation (Week 1)

1. **Extend existing test infrastructure**
   - Add round-trip validation functions to provider test utils
   - Create helper functions for importer execution
   - Implement workspace management for import testing

2. **Integrate with acceptance test framework**
   ```go
   // Add to provider/provider_test.go
   func testAccRoundTripPreCheck(t *testing.T) {
       testAccPreCheck(t) // Reuse existing precheck
       
       // Additional checks for round-trip testing
       if _, err := exec.LookPath("terraform"); err != nil {
           t.Fatal("terraform CLI not found")
       }
   }
   ```

3. **Create shared test configurations**
   - Extract common configs from existing tests
   - Build fixture library for different resource types
   - Ensure configs work for both provider and importer tests

### Phase 2: Core Testing (Week 2)

1. **Implement enhanced acceptance tests**
   ```go
   // provider/roundtrip_test.go
   func TestAccRoundTrip_BasicResources(t *testing.T) {
       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { testAccRoundTripPreCheck(t) },
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               // Step 1: Create resources
               {
                   Config: testAccBasicResourcesConfig(),
                   Check: resource.ComposeTestCheckFunc(
                       testAccCheckResourcesExist(),
                   ),
               },
               // Step 2: Standard import test
               {
                   ImportState:             true,
                   ImportStateVerify:       true,
                   ImportStateVerifyIgnore: computedFields,
                   ResourceName:            "polytomic_csv_connection.test",
               },
               // Step 3: Round-trip validation
               {
                   Config: testAccBasicResourcesConfig(),
                   Check: testAccRoundTripValidation(
                       []string{
                           "polytomic_csv_connection.test",
                           "polytomic_model.test",
                           "polytomic_sync.test",
                       },
                       RoundTripOptions{
                           IncludePermissions: false,
                       },
                   ),
               },
           },
       })
   }
   ```

2. **Create basic test fixtures**
   - Single connection resource
   - Connection → Model → Sync chain
   - Multiple independent resources

### Phase 3: Advanced Testing (Week 3)

1. **Handle special cases**
   
   **Sensitive Fields:**
   ```go
   func validateSensitiveField(t *testing.T, fieldName string, original, imported interface{}) {
       // Original should have value
       assert.NotEmpty(t, original)
       
       // Imported should reference variable
       assert.Contains(t, imported, "var.")
       
       // Variable should be declared
       assertVariableExists(t, fieldName)
   }
   ```
   
   **Resource References:**
   ```go
   func validateReference(t *testing.T, ref string) {
       // Parse reference: polytomic_connection.example.id
       parts := strings.Split(ref, ".")
       resourceType := parts[0]
       resourceName := parts[1]
       
       // Verify resource exists in imported config
       assertResourceExists(t, resourceType, resourceName)
   }
   ```

2. **Implement complex fixtures**
   - Resources with circular dependencies
   - Bulk operations (10+ resources)
   - Mixed resource types

### Phase 4: Validation & Reporting (Week 4)

1. **Enhanced validation logic**
   ```go
   type ValidationResult struct {
       Resource      string
       Field         string
       Original      interface{}
       Imported      interface{}
       Valid         bool
       SkipReason    string
       ErrorMessage  string
   }
   
   func generateValidationReport(results []ValidationResult) {
       // Create detailed HTML/JSON report
       // Include statistics and failure analysis
   }
   ```

2. **CI/CD Integration**
   ```yaml
   # .github/workflows/round-trip-tests.yml
   name: Round-Trip Tests
   on: [push, pull_request]
   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v2
         - uses: actions/setup-go@v2
         - uses: hashicorp/setup-terraform@v2
         - run: make test-roundtrip
   ```

## Field Handling Strategy

### Field Classification

| Field Type | Example | Import Behavior | Validation Strategy |
|------------|---------|-----------------|-------------------|
| **Identifiers** | `id`, `uuid` | Server-generated | Skip comparison |
| **Timestamps** | `created_at`, `updated_at` | Server-generated | Skip comparison |
| **Sensitive** | `password`, `api_key` | Variable reference | Verify variable exists |
| **References** | `connection_id` | Terraform reference | Validate reference syntax |
| **Optional** | `description` | Include if non-default | Compare if present in both |
| **Computed** | `status`, `version` | Read-only | Skip or verify read-only |
| **Arrays** | `fields`, `tags` | Preserve elements | Sort before comparison |
| **Maps** | `configuration` | Preserve keys | Deep comparison |

### Special Handling Rules

1. **Connection Resources**
   - Sensitive configuration fields → Variables
   - Provider-specific fields → Preserve structure
   - Optional OAuth tokens → Skip if absent

2. **Model Resources**
   - Field arrays → Maintain order
   - User-added fields → Separate handling
   - Tracking columns → Filter from export

3. **Sync Resources**
   - Schedule objects → Deep comparison
   - Mode-specific fields → Conditional validation
   - Identity resolution → Complex nested validation

## Test Scenarios

### Leveraging Existing Tests

The framework extends existing provider acceptance tests with round-trip validation:

| Existing Test | Round-Trip Enhancement | Coverage |
|--------------|----------------------|----------|
| `TestAccConnectionResource` | Add importer validation step | Single connection import/export |
| `TestAccModelResource` | Validate model with references | Model → Connection references |
| `TestAccSyncResource` | Full chain validation | Sync → Model → Connection |
| `TestAccPolicyResource` | Permission export testing | RBAC resources |

### New Round-Trip Specific Tests

1. **Basic Round-Trip Tests**
   ```go
   func TestAccRoundTrip_SingleConnection(t *testing.T)
   func TestAccRoundTrip_ModelWithSync(t *testing.T)
   func TestAccRoundTrip_BulkSync(t *testing.T)
   ```

2. **Complex Scenario Tests**
   ```go
   func TestAccRoundTrip_MultipleConnections(t *testing.T)
   func TestAccRoundTrip_ComplexDependencies(t *testing.T)
   func TestAccRoundTrip_WithPermissions(t *testing.T)
   ```

3. **Edge Case Tests**
   ```go
   func TestAccRoundTrip_SensitiveFields(t *testing.T)
   func TestAccRoundTrip_SpecialCharacters(t *testing.T)
   func TestAccRoundTrip_LargeConfiguration(t *testing.T)
   ```

## Success Metrics

### Required Outcomes
- ✅ 100% of basic scenarios pass without drift
- ✅ 95% of advanced scenarios pass (documented exceptions)
- ✅ No sensitive data exposed in generated configs
- ✅ All resource references maintain integrity
- ✅ Generated code passes terraform fmt/validate
- ✅ Import scripts execute without errors

### Performance Targets
- ⏱️ Basic scenario: < 30 seconds
- ⏱️ Complex scenario: < 2 minutes
- ⏱️ Full test suite: < 15 minutes

## Error Handling

### Expected Failures
1. **API Rate Limits**: Implement exponential backoff
2. **Transient Network**: Retry logic with timeout
3. **Resource Conflicts**: Cleanup and retry
4. **Import Errors**: Detailed logging and state recovery

### Running the Tests

### Environment Setup
```bash
# Required environment variables (same as provider tests)
export TF_ACC=1
export POLYTOMIC_API_KEY="your-api-key"
export POLYTOMIC_DEPLOYMENT_URL="https://app.polytomic.com"

# Optional: Test-specific overrides
export POLYTOMIC_TEST_ORG="test-organization"
export POLYTOMIC_TEST_SKIP_CLEANUP=false
```

### Test Execution
```bash
# Run all acceptance tests including round-trip
make testacc

# Run only round-trip tests
go test -v -run TestAccRoundTrip ./provider

# Run specific round-trip test
go test -v -run TestAccRoundTrip_BasicResources ./provider

# Run with detailed output
TF_LOG=DEBUG go test -v -run TestAccRoundTrip ./provider

# Keep test workspaces for debugging
TEST_KEEP_WORKSPACES=true go test -v -run TestAccRoundTrip ./provider
```

### CI/CD Integration
```yaml
# .github/workflows/acceptance-tests.yml
name: Acceptance Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - uses: hashicorp/setup-terraform@v2
      
      - name: Run Acceptance Tests
        env:
          TF_ACC: '1'
          POLYTOMIC_API_KEY: ${{ secrets.POLYTOMIC_API_KEY }}
          POLYTOMIC_DEPLOYMENT_URL: ${{ secrets.POLYTOMIC_URL }}
        run: |
          go test -v -timeout 30m -parallel 4 ./provider
```

## Key Advantages of This Approach

### 1. **Reusability**
- Leverages existing `terraform-plugin-testing` framework
- Shares test configurations between provider and importer tests
- Uses same test infrastructure (PreCheck, ProviderFactories)

### 2. **Maintainability**
- Tests live alongside provider tests
- Single test suite for both provider and importer
- Familiar patterns for Terraform developers

### 3. **Comprehensive Coverage**
- Built-in ImportState testing validates standard import
- Custom round-trip validation ensures exporter correctness
- Same fixtures test both creation and import paths

### 4. **CI/CD Ready**
- Integrates with existing test pipelines
- Uses standard `go test` tooling
- Parallel execution support

## Future Enhancements

### Phase 2 Features
1. **Extended ImportState Testing**: Leverage more ImportState features
   - `ImportStateId` and `ImportStateIdFunc` for complex imports
   - `ImportStateCheck` for custom validation logic
2. **State Migration Testing**: Validate provider version upgrades
3. **Performance Benchmarking**: Track import/export times

### Phase 3 Features
1. **Fuzzing with Acceptance Tests**: Generate random configs
2. **Multi-Provider Testing**: Test resources that span providers
3. **Drift Detection Reports**: Identify commonly drifting fields

## Implementation Timeline

| Week | Phase | Deliverables |
|------|-------|-------------|
| 1 | Foundation | Extend provider test infrastructure, add round-trip utils, integrate importer |
| 2 | Core Testing | Basic round-trip tests, ImportState validation, shared fixtures |
| 3 | Advanced Testing | Complex scenarios, sensitive field handling, dependency validation |
| 4 | CI/CD & Polish | CI pipeline integration, test reports, documentation |

## Appendix: Code Examples

### Complete Round-Trip Test Example
```go
// provider/roundtrip_connection_test.go
package provider

import (
    "testing"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoundTrip_PostgreSQLConnection(t *testing.T) {
    resourceName := "polytomic_postgresql_connection.test"
    
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccRoundTripPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:            testAccCheckConnectionDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create the connection
            {
                Config: testAccPostgreSQLConnectionConfig("roundtrip_test"),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(resourceName, "name", "roundtrip_test"),
                    resource.TestCheckResourceAttrSet(resourceName, "id"),
                ),
            },
            // Step 2: Test standard Terraform import
            {
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateVerifyIgnore: []string{
                    "configuration.password", // Sensitive field
                },
                ResourceName: resourceName,
            },
            // Step 3: Validate round-trip through importer
            {
                Config: testAccPostgreSQLConnectionConfig("roundtrip_test"),
                Check: testAccRoundTripValidation(
                    []string{resourceName},
                    RoundTripOptions{
                        ValidateSensitive: true,
                        IgnoreFields: []string{
                            "created_at",
                            "updated_at",
                        },
                    },
                ),
            },
        },
    })
}

func testAccPostgreSQLConnectionConfig(name string) string {
    return fmt.Sprintf(`
resource "polytomic_postgresql_connection" "test" {
    name = "%s"
    configuration = {
        hostname = "localhost"
        port     = 5432
        database = "testdb"
        username = "testuser"
        password = "testpass"
    }
}`, name)
}
```

### Field Validation Example
```go
func compareSchedule(original, imported interface{}) error {
    origSchedule := original.(map[string]interface{})
    impSchedule := imported.(map[string]interface{})
    
    // Compare frequency
    if origSchedule["frequency"] != impSchedule["frequency"] {
        return fmt.Errorf("frequency mismatch: %v != %v", 
            origSchedule["frequency"], impSchedule["frequency"])
    }
    
    // Compare day_of_week array (order doesn't matter)
    origDays := toStringSlice(origSchedule["day_of_week"])
    impDays := toStringSlice(impSchedule["day_of_week"])
    sort.Strings(origDays)
    sort.Strings(impDays)
    
    if !reflect.DeepEqual(origDays, impDays) {
        return fmt.Errorf("day_of_week mismatch")
    }
    
    return nil
}
```

### Example Validation Report
```json
{
  "test_name": "ComplexDependencies",
  "duration": "45.2s",
  "total_resources": 25,
  "validated_fields": 312,
  "skipped_fields": 48,
  "failures": 0,
  "details": [
    {
      "resource": "polytomic_model.customer",
      "fields_validated": 15,
      "fields_skipped": 3,
      "skip_reasons": {
        "id": "server_generated",
        "created_at": "timestamp",
        "updated_at": "timestamp"
      }
    }
  ]
}
```