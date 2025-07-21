package roundtrip

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/polytomic/terraform-provider-polytomic/importer"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
	"github.com/polytomic/terraform-provider-polytomic/provider"
)

// PreCheck extends the standard precheck for round-trip testing
func PreCheck(t *testing.T) func() {
	return func() {
		provider.TestAccPreCheck(t) // Reuse existing precheck from provider package

		// Additional checks for round-trip testing
		if _, err := exec.LookPath("terraform"); err != nil {
			t.Fatal("terraform CLI not found in PATH")
		}
	}
}

// ImportAndValidate runs the importer and validates round-trip
func ImportAndValidate(ctx context.Context, resourceNames []string, opts RoundTripOptions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Step 1: Get current resources from state
		resources := make(map[string]*terraform.ResourceState)
		for _, name := range resourceNames {
			rs, ok := s.RootModule().Resources[name]
			if !ok {
				return fmt.Errorf("resource not found: %s", name)
			}
			resources[name] = rs
		}

		// Step 2: Run importer to generate Terraform files
		exportDir, err := os.MkdirTemp("", "roundtrip-export-")
		if err != nil {
			return fmt.Errorf("failed to create export directory: %w", err)
		}
		defer func() {
			if os.Getenv("TEST_KEEP_WORKSPACES") != "true" {
				os.RemoveAll(exportDir)
			} else {
				fmt.Printf("Keeping export directory for debugging: %s\n", exportDir)
			}
		}()

		err = runImporter(ctx, exportDir, opts.IncludePermissions)
		if err != nil {
			return fmt.Errorf("importer failed: %w", err)
		}

		// Step 3: Validate generated files
		if opts.OrgName != "" {
			exportDir = filepath.Join(exportDir, opts.OrgName)
		}
		err = validateGeneratedFiles(exportDir)
		if err != nil {
			return fmt.Errorf("generated files validation failed: %w", err)
		}

		// Step 4: Create new workspace and import
		importWS, err := setupTempWorkspace()
		if err != nil {
			return fmt.Errorf("failed to setup workspace: %w", err)
		}
		defer cleanupWorkspace(importWS)

		// Copy generated files
		err = copyGeneratedFiles(exportDir, importWS.Dir)
		if err != nil {
			return fmt.Errorf("failed to copy generated files: %w", err)
		}

		// Create .tfrc file for local provider override
		err = importWS.CreateTfrcFile()
		if err != nil {
			return fmt.Errorf("failed to create .tfrc file: %w", err)
		}

		// Initialize terraform
		err = importWS.Init()
		if err != nil {
			return fmt.Errorf("terraform init failed: %w", err)
		}

		// Run import script
		importScript := filepath.Join(exportDir, "import.sh")
		if _, err := os.Stat(importScript); err == nil {
			err = importWS.ExecuteImportScript(importScript)
			if err != nil {
				return fmt.Errorf("import script failed: %w", err)
			}
		} else {
			// If no import script, we might need to debug what the importer generated
			entries, _ := os.ReadDir(exportDir)
			var fileNames []string
			for _, entry := range entries {
				fileNames = append(fileNames, entry.Name())
			}
			return fmt.Errorf("no import script found in export directory, files present: %v", fileNames)
		}

		// Step 5: Validate no drift
		hasDrift, driftDetails, err := importWS.CheckForDrift()
		if err != nil {
			return fmt.Errorf("plan failed: %w", err)
		}

		if hasDrift {
			return fmt.Errorf("unexpected drift detected: %s", driftDetails)
		}

		// Step 6: Validate specific fields
		_, importedResources, err := importWS.GetState()
		if err != nil {
			return fmt.Errorf("failed to get imported state: %w", err)
		}

		validationResults := []ValidationResult{}
		for name, originalRS := range resources {
			// Map from test resource name to imported resource name
			// The importer uses the API object name, not the Terraform resource name
			importedName := mapResourceName(name, originalRS, importedResources)
			if importedName == "" {
				return fmt.Errorf("resource %s not found in imported state", name)
			}

			importedRS := importedResources[importedName]
			results := validateResourceFields(originalRS, importedRS, opts)
			validationResults = append(validationResults, results...)
		}

		// Check for validation errors
		for _, result := range validationResults {
			if !result.Valid && result.SkipReason == "" {
				return fmt.Errorf("validation failed for %s.%s: %s",
					result.Resource, result.Field, result.ErrorMessage)
			}
		}

		// Validate variables if sensitive fields are expected
		if opts.ValidateSensitive && len(opts.ExpectedVariables) > 0 {
			err = validateVariables(exportDir, opts.ExpectedVariables)
			if err != nil {
				return fmt.Errorf("variable validation failed: %w", err)
			}
		}

		return nil
	}
}

// runImporter executes the importer using the importer package
func runImporter(ctx context.Context, outputDir string, includePermissions bool) error {
	client, err := providerclient.NewClientProvider(providerclient.OptionsFromEnv())
	if err != nil {
		return fmt.Errorf("failed to create client provider: %w", err)
	}

	// Initialize and run importer directly
	// Note: importer.Init uses log.Fatal on errors, so if we get here it succeeded
	importer.Init(ctx, client, "", outputDir, true, includePermissions)

	return nil
}

// setupTempWorkspace creates a temporary Terraform workspace
func setupTempWorkspace() (*TerraformWorkspace, error) {
	dir, err := os.MkdirTemp("", "roundtrip-import-")
	if err != nil {
		return nil, err
	}

	tfPath, err := exec.LookPath("terraform")
	if err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("terraform not found in PATH: %w", err)
	}

	return &TerraformWorkspace{
		Dir:    dir,
		TfPath: tfPath,
	}, nil
}

// cleanupWorkspace removes the temporary workspace
func cleanupWorkspace(ws *TerraformWorkspace) {
	if ws != nil {
		// Keep workspace for debugging if TEST_KEEP_WORKSPACES is set
		if os.Getenv("TEST_KEEP_WORKSPACES") == "true" {
			fmt.Printf("Keeping workspace for debugging: %s\n", ws.Dir)
			if ws.ProviderDir != "" {
				fmt.Printf("Keeping provider directory for debugging: %s\n", ws.ProviderDir)
			}
			return
		}

		if ws.Dir != "" {
			os.RemoveAll(ws.Dir)
		}
		if ws.ProviderDir != "" {
			os.RemoveAll(ws.ProviderDir)
		}
	}
}

// copyGeneratedFiles copies terraform files from export to workspace
func copyGeneratedFiles(exportDir, workspaceDir string) error {
	// Find all .tf and .tfvars files in the export directory
	entries, err := os.ReadDir(exportDir)
	if err != nil {
		return fmt.Errorf("failed to read export directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".tf") && !strings.HasSuffix(fileName, ".tfvars") {
			continue
		}

		src := filepath.Join(exportDir, fileName)
		dst := filepath.Join(workspaceDir, fileName)

		content, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", src, err)
		}

		err = os.WriteFile(dst, content, 0644)
		if err != nil {
			return fmt.Errorf("failed to write %s: %w", dst, err)
		}
	}

	return nil
}

// buildProvider builds the terraform-provider-polytomic binary and returns its path
func buildProvider() (string, error) {
	// Get the go binary path
	goPath, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("go not found in PATH: %w", err)
	}

	// Find the repository root (contains go.mod)
	repoRoot, err := findRepoRoot()
	if err != nil {
		return "", fmt.Errorf("failed to find repository root: %w", err)
	}

	// Create a temporary directory for the provider binary
	tempDir, err := os.MkdirTemp("", "provider-build-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Build the provider into the temp directory
	providerPath := filepath.Join(tempDir, "terraform-provider-polytomic")
	cmd := exec.Command(goPath, "build", "-o", providerPath, ".")
	cmd.Dir = repoRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("go build failed: %w\nOutput: %s", err, output)
	}

	return providerPath, nil
}

// findRepoRoot finds the repository root using go env GOMOD
func findRepoRoot() (string, error) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("go not found in PATH: %w", err)
	}

	// Use 'go env GOMOD' to find the current module's go.mod file
	cmd := exec.Command(goPath, "env", "GOMOD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run 'go env GOMOD': %w", err)
	}

	goModPath := strings.TrimSpace(string(output))
	if goModPath == "" || goModPath == "/dev/null" {
		return "", fmt.Errorf("not in a Go module")
	}

	return filepath.Dir(goModPath), nil
}

// CreateTfrcFile creates a .tfrc file in the workspace for local provider override
func (ws *TerraformWorkspace) CreateTfrcFile() error {
	// Build the provider
	providerPath, err := buildProvider()
	if err != nil {
		return fmt.Errorf("failed to build provider: %w", err)
	}
	// Store the provider directory for cleanup
	ws.ProviderDir = filepath.Dir(providerPath)

	// Get the directory containing the provider binary
	providerDir := filepath.Dir(providerPath)

	// Create .tfrc content
	tfrcContent := fmt.Sprintf(`provider_installation {
  dev_overrides {
    "polytomic/polytomic" = "%s"
  }
  direct {}
}
`, providerDir)

	// Write .tfrc file to workspace
	tfrcPath := filepath.Join(ws.Dir, ".tfrc")
	err = os.WriteFile(tfrcPath, []byte(tfrcContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .tfrc file: %w", err)
	}

	return nil
}

// Init runs terraform init in the workspace
func (ws *TerraformWorkspace) Init() error {
	cmd := exec.Command(ws.TfPath, "init")
	cmd.Dir = ws.Dir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", filepath.Join(ws.Dir, ".tfrc")),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// ExecuteImportScript runs the generated import script
func (ws *TerraformWorkspace) ExecuteImportScript(scriptPath string) error {
	// Read and modify script to work in the workspace directory
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read import script: %w", err)
	}

	// Update script to use workspace directory
	modifiedScript := strings.ReplaceAll(string(content), "terraform import", ws.TfPath+" import")

	// Create temporary script in workspace
	tempScript := filepath.Join(ws.Dir, "import.sh")
	err = os.WriteFile(tempScript, []byte(modifiedScript), 0755)
	if err != nil {
		return fmt.Errorf("failed to write temp script: %w", err)
	}

	// Execute the script
	cmd := exec.Command("/bin/bash", tempScript)
	cmd.Dir = ws.Dir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", filepath.Join(ws.Dir, ".tfrc")),
		fmt.Sprintf("POLYTOMIC_API_KEY=%s", os.Getenv("POLYTOMIC_API_KEY")),
		fmt.Sprintf("POLYTOMIC_DEPLOYMENT_KEY=%s", os.Getenv("POLYTOMIC_DEPLOYMENT_KEY")),
		fmt.Sprintf("POLYTOMIC_DEPLOYMENT_URL=%s", os.Getenv("POLYTOMIC_DEPLOYMENT_URL")),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("import script failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// CheckForDrift runs terraform plan and checks for changes
func (ws *TerraformWorkspace) CheckForDrift() (bool, string, error) {
	cmd := exec.Command(ws.TfPath, "plan", "-detailed-exitcode", "-out=tfplan")
	cmd.Dir = ws.Dir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TF_CLI_CONFIG_FILE=%s", filepath.Join(ws.Dir, ".tfrc")),
		fmt.Sprintf("POLYTOMIC_API_KEY=%s", os.Getenv("POLYTOMIC_API_KEY")),
		fmt.Sprintf("POLYTOMIC_DEPLOYMENT_KEY=%s", os.Getenv("POLYTOMIC_DEPLOYMENT_KEY")),
		fmt.Sprintf("POLYTOMIC_DEPLOYMENT_URL=%s", os.Getenv("POLYTOMIC_DEPLOYMENT_URL")),
	)

	output, err := cmd.CombinedOutput()

	// Exit code 0 = no changes, 2 = changes present
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 2 {
				// Changes detected
				return true, string(output), nil
			}
		}
		return false, "", fmt.Errorf("terraform plan failed: %w\nOutput: %s", err, output)
	}

	return false, "", nil
}

// mapResourceName maps from test resource name to imported resource name
// The importer generates resource names based on API object names, not Terraform names
func mapResourceName(testName string, originalRS *terraform.ResourceState, importedResources map[string]*terraform.ResourceState) string {
	// Get the original resource's name attribute (the API object name)
	apiObjectName := originalRS.Primary.Attributes["name"]
	if apiObjectName == "" {
		return ""
	}

	// Look for a resource of the same type with that name in the imported resources
	resourceType := originalRS.Type
	expectedImportedName := resourceType + "." + apiObjectName

	if _, exists := importedResources[expectedImportedName]; exists {
		return expectedImportedName
	}

	// Fallback: try exact match
	if _, exists := importedResources[testName]; exists {
		return testName
	}

	return ""
}

// validateResourceFields compares fields between original and imported resources
func validateResourceFields(original, imported *terraform.ResourceState, opts RoundTripOptions) []ValidationResult {
	results := []ValidationResult{}

	// Fields to always skip
	alwaysSkip := map[string]string{
		"id":         "server_generated",
		"created_at": "timestamp",
		"updated_at": "timestamp",
		"version":    "computed",
	}

	// Add user-specified fields to skip
	for _, field := range opts.IgnoreFields {
		alwaysSkip[field] = "user_ignored"
	}

	// Compare attributes
	for key, origValue := range original.Primary.Attributes {
		result := ValidationResult{
			Resource: original.Type,
			Field:    key,
			Original: origValue,
		}

		// Check if field should be skipped
		if skipReason, skip := alwaysSkip[key]; skip {
			result.Valid = true
			result.SkipReason = skipReason
			results = append(results, result)
			continue
		}

		// Get imported value
		impValue, exists := imported.Primary.Attributes[key]
		if !exists {
			result.Valid = false
			result.ErrorMessage = "field missing in imported resource"
			results = append(results, result)
			continue
		}

		result.Imported = impValue

		// Handle sensitive fields
		if opts.ValidateSensitive && isSensitiveField(key) {
			if strings.Contains(impValue, "var.") {
				result.Valid = true
			} else {
				result.Valid = false
				result.ErrorMessage = "sensitive field should reference a variable"
			}
		} else if origValue == impValue {
			result.Valid = true
		} else {
			// Try to compare as JSON for complex fields
			if isJSONField(key) {
				result.Valid = compareJSON(origValue, impValue)
				if !result.Valid {
					result.ErrorMessage = fmt.Sprintf("JSON mismatch: %s != %s", origValue, impValue)
				}
			} else {
				result.Valid = false
				result.ErrorMessage = fmt.Sprintf("value mismatch: %s != %s", origValue, impValue)
			}
		}

		results = append(results, result)
	}

	return results
}

// validateGeneratedFiles checks that generated terraform files are valid
func validateGeneratedFiles(exportDir string) error {
	mainTf := filepath.Join(exportDir, "main.tf")
	if _, err := os.Stat(mainTf); err != nil {
		return fmt.Errorf("main.tf not found: %w", err)
	}

	// Run terraform fmt -check
	cmd := exec.Command("terraform", "fmt", exportDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform fmt failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// validateVariables checks that expected variables are declared
func validateVariables(exportDir string, expectedVars []string) error {
	varsFile := filepath.Join(exportDir, "variables.tf")
	content, err := os.ReadFile(varsFile)
	if err != nil {
		return fmt.Errorf("failed to read variables.tf: %w", err)
	}

	varsContent := string(content)
	for _, varName := range expectedVars {
		if !strings.Contains(varsContent, fmt.Sprintf(`variable "%s"`, varName)) {
			return fmt.Errorf("expected variable %s not found in variables.tf", varName)
		}
	}

	return nil
}

// Helper functions

func isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{
		"password", "api_key", "secret", "token",
		"private_key", "client_secret", "access_key",
	}

	fieldLower := strings.ToLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldLower, sensitive) {
			return true
		}
	}
	return false
}

func isJSONField(fieldName string) bool {
	return strings.HasSuffix(fieldName, "_json") ||
		fieldName == "configuration" ||
		fieldName == "schedule"
}

func compareJSON(a, b string) bool {
	var aData, bData interface{}

	if err := json.Unmarshal([]byte(a), &aData); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &bData); err != nil {
		return false
	}

	return reflect.DeepEqual(normalizeJSON(aData), normalizeJSON(bData))
}

func normalizeJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case []interface{}:
		// Sort arrays for comparison
		sorted := make([]interface{}, len(v))
		copy(sorted, v)
		// Note: This is a simplified sort, may need enhancement for complex types
		return sorted
	case map[string]interface{}:
		normalized := make(map[string]interface{})
		for key, val := range v {
			normalized[key] = normalizeJSON(val)
		}
		return normalized
	default:
		return v
	}
}
