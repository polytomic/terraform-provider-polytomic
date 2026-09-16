package importer

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// TestSyncsTargetFilters verifies that target filters are exported as
// target_filters, and that syncs whose filters compare against a model field,
// which the provider cannot express, are skipped.
func TestSyncsTargetFilters(t *testing.T) {
	const sync = `{"data":{"id":"%s","name":"%s","active":false,"mode":"update",
		"schedule":{"frequency":"manual"},
		"target":{"connection_id":"conn-1","object":"contacts","filter_logic":"A"},
		"filters":[%s]}}`
	s := NewSyncs(newTestClient(t, map[string]string{
		"/api/syncs": `{"data":[{"id":"sync-value","name":"Value filter"},{"id":"sync-field","name":"Field filter"}]}`,
		"/api/syncs/sync-value": fmt.Sprintf(sync, "sync-value", "Value filter",
			`{"field":{"model_id":"00000000-0000-0000-0000-000000000000","field":""},
			  "field_id":"lifecyclestage","field_type":"Target","function":"Equality","value":"lead","label":"A"}`),
		"/api/syncs/sync-field": fmt.Sprintf(sync, "sync-field", "Field filter",
			`{"field":{"model_id":"00000000-0000-0000-0000-000000000000","field":""},
			  "field_id":"lastname","field_type":"Target","function":"Equality",
			  "value_field":{"model_id":"model-1","field":"id"},"label":"A"}`),
	}))
	if err := s.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := sortedKeys(s.Resources); len(got) != 1 || got[0] != "value_filter" {
		t.Fatalf("resources: got %v, want [value_filter]", got)
	}

	var tf bytes.Buffer
	if err := s.GenerateTerraformFiles(context.Background(), &tf, nil); err != nil {
		t.Fatal(err)
	}
	if _, diags := hclsyntax.ParseConfig(tf.Bytes(), SyncResourceFileName, hcl.InitialPos); diags.HasErrors() {
		t.Fatalf("%s\n%s", diags, tf.String())
	}
	for _, want := range []string{`target_filters\s*=`, `field\s*=\s*"lifecyclestage"`, `value\s*=\s*jsonencode\(\s*"lead"\)`} {
		if !regexp.MustCompile(want).Match(tf.Bytes()) {
			t.Errorf("missing %s in:\n%s", want, tf.String())
		}
	}
	if regexp.MustCompile(`(?m)^\s*filters\s*=`).Match(tf.Bytes()) {
		t.Errorf("target filter exported as a model filter:\n%s", tf.String())
	}
}

// TestBulkSyncsSkipsMissingDefaultSchedule verifies that a bulk sync without a
// default schedule, which the provider requires, is not exported.
func TestBulkSyncsSkipsMissingDefaultSchedule(t *testing.T) {
	b := NewBulkSyncs(newTestClient(t, map[string]string{
		"/api/bulk/syncs": `{"data":[
			{"id":"aaaaaaaa-0000-4000-8000-000000000001","name":"Scheduled","default_schedule":{"frequency":"manual"}},
			{"id":"bbbbbbbb-0000-4000-8000-000000000002","name":"Unscheduled"}
		]}`,
	}))
	if err := b.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := sortedKeys(b.Resources); len(got) != 1 || got[0] != "scheduled_aaaaaaaa" {
		t.Errorf("resources: got %v, want [scheduled_aaaaaaaa]", got)
	}
}
