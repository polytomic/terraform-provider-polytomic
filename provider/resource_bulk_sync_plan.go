package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithModifyPlan = (*bulkSyncResource)(nil)

// Terraform cannot correlate nested set elements when omitted computed
// attributes change their identity. The framework then marks all computed
// values unknown, including audit fields, even when no settings changed.
// Restore refreshed state only if every configured value still agrees with it.
// Real changes keep the normal plan, so outputs and audit fields can change.
func (r *bulkSyncResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	var config, plan, state types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if bulkSyncPlanObjectMatchesState(config, plan, state, schemaResp.Schema.Attributes) {
		resp.Plan.Raw = req.State.Raw
	}
}

func bulkSyncPlanObjectMatchesState(config, plan, state types.Object, attributes map[string]schema.Attribute) bool {
	if config.IsNull() || config.IsUnknown() || plan.IsNull() || plan.IsUnknown() || state.IsNull() || state.IsUnknown() {
		return false
	}
	configAttrs, planAttrs, stateAttrs := config.Attributes(), plan.Attributes(), state.Attributes()
	for name, attribute := range attributes {
		c, p, s := configAttrs[name], planAttrs[name], stateAttrs[name]
		// Only unknowns originating from omitted computed attributes can stand
		// for the current value. Explicit unknown configuration must stay unknown.
		if attribute.IsComputed() && c.IsNull() && p.IsUnknown() {
			continue
		}
		if c.IsUnknown() || p.IsUnknown() {
			return false
		}
		if p.Equal(s) {
			continue
		}
		if c.IsNull() || p.IsNull() || s.IsNull() {
			return false
		}
		switch a := attribute.(type) {
		case schema.SingleNestedAttribute:
			if !bulkSyncPlanObjectMatchesState(c.(types.Object), p.(types.Object), s.(types.Object), a.Attributes) {
				return false
			}
		case schema.SetNestedAttribute:
			if !bulkSyncPlanSetMatchesState(c.(types.Set), p.(types.Set), s.(types.Set), a.NestedObject.Attributes) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// Schema and field set order is unstable; IDs are their configuration identity.
func bulkSyncPlanSetMatchesState(config, plan, state types.Set, attributes map[string]schema.Attribute) bool {
	if len(config.Elements()) != len(state.Elements()) || len(plan.Elements()) != len(state.Elements()) {
		return false
	}
	index := func(set types.Set) map[string]types.Object {
		result := make(map[string]types.Object)
		for _, element := range set.Elements() {
			object, ok := element.(types.Object)
			if !ok || object.IsNull() || object.IsUnknown() {
				return nil
			}
			id, ok := object.Attributes()["id"].(types.String)
			if !ok || id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
				return nil
			}
			if _, exists := result[id.ValueString()]; exists {
				return nil
			}
			result[id.ValueString()] = object
		}
		return result
	}
	configs, plans, states := index(config), index(plan), index(state)
	if configs == nil || plans == nil || states == nil {
		return false
	}
	for id, c := range configs {
		p, ok := plans[id]
		if !ok {
			return false
		}
		s, ok := states[id]
		if !ok {
			return false
		}
		if !bulkSyncPlanObjectMatchesState(c, p, s, attributes) {
			return false
		}
	}
	return true
}
