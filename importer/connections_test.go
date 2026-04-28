package importer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// TestSubstituteVarRefs verifies that varRefSentinel placeholders are
// rewritten to bare var.<name> traversals after HCL rendering.
func TestSubstituteVarRefs(t *testing.T) {
	in := []byte(`api_key = "__VARREF_posthog_prod_api_key__"
location = "us"
nested = "__VARREF_other_field__"
`)
	got := string(substituteVarRefs(in))

	if !strings.Contains(got, "api_key = var.posthog_prod_api_key\n") {
		t.Errorf("api_key not substituted: %q", got)
	}
	if !strings.Contains(got, "nested = var.other_field\n") {
		t.Errorf("nested not substituted: %q", got)
	}
	if strings.Contains(got, "__VARREF_") {
		t.Errorf("placeholder leaked: %q", got)
	}
	if !strings.Contains(got, `location = "us"`) {
		t.Errorf("non-placeholder string was rewritten: %q", got)
	}
}

// TestRenderConnectionWithVarRef exercises the full HCL rendering path:
// build a configuration with a sentinel value, render it through
// hclwrite, and confirm the post-processed output references a variable
// rather than a string literal.
func TestRenderConnectionWithVarRef(t *testing.T) {
	conn := Connection{
		ID:       pointer.ToString("conn-id"),
		Resource: "polytomic_posthog_connection",
		Name:     pointer.ToString("Prod"),
		Configuration: map[string]interface{}{
			"api_key":  varRefSentinel("posthog_prod_api_key"),
			"location": "us",
			"project":  "12345",
		},
	}

	hclFile := hclwrite.NewEmptyFile()
	body := hclFile.Body()
	resourceBlock := body.AppendNewBlock("resource", []string{conn.Resource, "posthog_prod"})
	resourceBlock.Body().SetAttributeValue("name", cty.StringVal(*conn.Name))
	resourceBlock.Body().SetAttributeTraversal("organization",
		hcl.Traversal{
			hcl.TraverseRoot{Name: "local"},
			hcl.TraverseAttr{Name: "organization_id"},
		},
	)
	resourceBlock.Body().SetAttributeValue("configuration", typeConverter(conn.Configuration))

	out := substituteVarRefs(hclFile.Bytes())

	want := []string{
		"api_key  = var.posthog_prod_api_key",
		`location = "us"`,
		`project  = "12345"`,
		"organization = local.organization_id",
	}
	for _, w := range want {
		if !bytes.Contains(out, []byte(w)) {
			t.Errorf("missing %q in output:\n%s", w, out)
		}
	}
	if bytes.Contains(out, []byte("__VARREF_")) {
		t.Errorf("placeholder leaked in output:\n%s", out)
	}
}
