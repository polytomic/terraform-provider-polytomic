package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// MergeSetElements merges a set of objects from the plan with API response data.
// It preserves user-specified values from planSet and populates unknown/computed
// fields from apiData based on a matching key.
//
// This is useful when the API returns more fields than the user specified,
// preventing "inconsistent result after apply" errors while ensuring all
// computed fields have known values.
//
// Parameters:
//   - ctx: context for the operation
//   - planSet: The set from the plan containing user-specified values
//   - apiData: Slice of structs from the API response
//   - elementType: The Terraform object type for set elements
//   - getKey: Function to extract the matching key from a plan element
//   - getAPIKey: Function to extract the matching key from an API element
//   - mergeFunc: Function to merge a plan element with matching API element
//
// Returns the merged set and any diagnostics.
func MergeSetElements[P any, A any](
	ctx context.Context,
	planSet basetypes.SetValue,
	apiData []A,
	elementType types.ObjectType,
	getKey func(P) string,
	getAPIKey func(A) string,
	mergeFunc func(context.Context, P, A) (P, diag.Diagnostics),
) (basetypes.SetValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Extract plan elements
	var planElements []P
	diags = planSet.ElementsAs(ctx, &planElements, false)
	if diags.HasError() {
		return types.SetNull(elementType), diags
	}

	// Create a map of API data by key for lookup
	apiMap := make(map[string]A)
	for _, apiItem := range apiData {
		apiMap[getAPIKey(apiItem)] = apiItem
	}

	// Merge plan elements with API data
	mergedElements := make([]P, 0, len(planElements))
	for _, planElement := range planElements {
		merged := planElement
		key := getKey(planElement)

		// Look up corresponding API data
		if apiItem, ok := apiMap[key]; ok {
			var mergeDiags diag.Diagnostics
			merged, mergeDiags = mergeFunc(ctx, planElement, apiItem)
			diags.Append(mergeDiags...)
			if mergeDiags.HasError() {
				return types.SetNull(elementType), diags
			}
		}

		mergedElements = append(mergedElements, merged)
	}

	// Convert merged elements back to a set
	result, resultDiags := types.SetValueFrom(ctx, elementType, mergedElements)
	diags.Append(resultDiags...)
	return result, diags
}

// PopulateUnknownString sets a string field to the provided value if it's currently unknown or null.
// This is useful for populating computed fields while preserving user-specified values.
func PopulateUnknownString(field basetypes.StringValue, value *string) basetypes.StringValue {
	if field.IsUnknown() || field.IsNull() {
		return types.StringPointerValue(value)
	}
	return field
}

// PopulateUnknownBool sets a bool field to the provided value if it's currently unknown.
// This is useful for populating computed fields while preserving user-specified values.
func PopulateUnknownBool(field basetypes.BoolValue, value *bool) basetypes.BoolValue {
	if field.IsUnknown() {
		return types.BoolPointerValue(value)
	}
	return field
}

// PopulateUnknownSet sets a set field from API data if it's currently unknown.
// This is useful for populating computed collection fields.
func PopulateUnknownSet[T any](
	ctx context.Context,
	field basetypes.SetValue,
	apiData []T,
	elementType attr.Type,
) (basetypes.SetValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !field.IsUnknown() {
		return field, diags
	}

	if len(apiData) > 0 {
		result, resultDiags := types.SetValueFrom(ctx, elementType, apiData)
		diags.Append(resultDiags...)
		return result, diags
	}

	return types.SetNull(elementType), diags
}
