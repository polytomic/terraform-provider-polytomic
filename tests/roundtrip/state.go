package roundtrip

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// StateResource represents a resource in the newer state format
type StateResource struct {
	Mode      string                  `json:"mode"`
	Type      string                  `json:"type"`
	Name      string                  `json:"name"`
	Provider  string                  `json:"provider"`
	Instances []StateResourceInstance `json:"instances"`
}

type StateResourceInstance struct {
	Attributes map[string]interface{} `json:"attributes"`
}

// RawState represents the raw state file structure to handle both old and new formats
type RawState struct {
	Version   int             `json:"version"`
	Resources []StateResource `json:"resources,omitempty"`
}

// getWorkspaceState reads the terraform state from workspace
func getWorkspaceState(ws *TerraformWorkspace) (*terraform.State, map[string]*terraform.ResourceState, error) {
	stateFile := filepath.Join(ws.Dir, "terraform.tfstate")

	content, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// First try to parse as the testing library's expected format
	var state terraform.State
	err = json.Unmarshal(content, &state)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Also parse raw to handle new format
	var rawState RawState
	err = json.Unmarshal(content, &rawState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse raw state: %w", err)
	}

	// Extract resources using helper
	resources := getStateResources(&state, &rawState)

	return &state, resources, nil
}

// getStateResources extracts resources from state regardless of format (v3 modules vs v4 resources)
func getStateResources(state *terraform.State, rawState *RawState) map[string]*terraform.ResourceState {
	resources := make(map[string]*terraform.ResourceState)

	// Try new format first (state version 4+)
	if len(rawState.Resources) > 0 {
		for _, resource := range rawState.Resources {
			if resource.Mode == "managed" && len(resource.Instances) > 0 {
				key := resource.Type + "." + resource.Name
				// Convert attributes map to the expected format
				attrs := make(map[string]string)
				for k, v := range resource.Instances[0].Attributes {
					// Handle complex objects by flattening them
					if v == nil {
						attrs[k] = ""
					} else {
						switch val := v.(type) {
						case string:
							attrs[k] = val
						case map[string]interface{}:
							// Flatten nested objects to match Terraform's internal representation
							flattenAttributes(attrs, k, val)
						case []interface{}:
							// Handle arrays - set count and flatten elements
							attrs[k+".#"] = fmt.Sprintf("%d", len(val))
							for i, item := range val {
								flattenAttributes(attrs, fmt.Sprintf("%s.%d", k, i), item)
							}
						default:
							attrs[k] = fmt.Sprintf("%v", v)
						}
					}
				}

				// Add the top-level count attribute
				attrs["%"] = fmt.Sprintf("%d", len(resource.Instances[0].Attributes))

				resources[key] = &terraform.ResourceState{
					Type:     resource.Type,
					Provider: resource.Provider,
					Primary: &terraform.InstanceState{
						Attributes: attrs,
					},
				}
			}
		}
		return resources
	}

	// Fall back to old format (modules)
	if state.RootModule() != nil {
		return state.RootModule().Resources
	}

	return resources
}

// flattenAttributes recursively flattens nested objects to match Terraform's internal representation
func flattenAttributes(attrs map[string]string, prefix string, value interface{}) {
	switch val := value.(type) {
	case map[string]interface{}:
		// Set the count for this map
		attrs[prefix+".%"] = fmt.Sprintf("%d", len(val))
		// Flatten each key-value pair
		for k, v := range val {
			flattenAttributes(attrs, prefix+"."+k, v)
		}
	case []interface{}:
		// Set the count for this array
		attrs[prefix+".#"] = fmt.Sprintf("%d", len(val))
		// Flatten each array element
		for i, item := range val {
			flattenAttributes(attrs, fmt.Sprintf("%s.%d", prefix, i), item)
		}
	case nil:
		attrs[prefix] = ""
	default:
		attrs[prefix] = fmt.Sprintf("%v", val)
	}
}
