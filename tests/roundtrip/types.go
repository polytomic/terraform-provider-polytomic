package roundtrip

// RoundTripOptions configures round-trip validation behavior
type RoundTripOptions struct {
	IncludePermissions bool
	ValidateSensitive  bool
	IgnoreFields       []string
	ExpectedVariables  []string
	// OrgName contains the name of the organization being tested; this will be
	// empty if tests are running with an API key.
	OrgName string
}

// ValidationResult represents the result of comparing a field between original and imported resources
type ValidationResult struct {
	Resource     string
	Field        string
	Original     interface{}
	Imported     interface{}
	Valid        bool
	SkipReason   string
	ErrorMessage string
}

// TerraformWorkspace represents a temporary terraform workspace for testing
type TerraformWorkspace struct {
	Dir         string
	TfPath      string
	ProviderDir string // Directory containing the built provider binary (for cleanup)
}
