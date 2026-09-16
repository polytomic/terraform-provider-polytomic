package importer

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ptclient "github.com/polytomic/polytomic-go/v25/client"
	"github.com/polytomic/polytomic-go/v25/option"
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

// newTestClient returns a client for a server that responds to each path in
// responses with its JSON body, and to any other path with 404.
func newTestClient(t *testing.T, responses map[string]string) *ptclient.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := responses[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return ptclient.NewClient(option.WithBaseURL(srv.URL), option.WithToken("token"), option.WithMaxAttempts(1))
}

// TestConnectionsInitGeneratesValidConfig covers connections whose stored
// configuration does not satisfy the provider schema as-is.
func TestConnectionsInitGeneratesValidConfig(t *testing.T) {
	c := NewConnections(newTestClient(t, map[string]string{
		"/api/connections": `{"data":[
			{"id":"conn-apple","name":"Apple Ads","organization_id":"org-1","type":{"id":"apple_ads"},
			 "configuration":{"client_id":"c","key_id":"k","team_id":"t"}},
			{"id":"conn-meta","name":"Metadata","organization_id":"org-1","type":{"id":"polytomic_metadata"},
			 "configuration":{"auth_mode":"deployment_api_key","deployment_api_key":"**********"}},
			{"id":"conn-attr","name":"Attribution","organization_id":"org-1","type":{"id":"apple_ads_attribution"},
			 "configuration":{}}
		]}`,
	}))
	if err := c.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	// A required field the connection predates becomes an input variable.
	apple, ok := c.Resources["apple_ads"]
	if !ok {
		t.Fatalf("Apple Ads connection was skipped: %v", sortedKeys(c.Resources))
	}
	if got := apple.Configuration.(map[string]any)["org_id"]; got != varRefSentinel("apple_ads_org_id") {
		t.Errorf("org_id: got %v", got)
	}
	wantVars := []Variable{{Name: "apple_ads_org_id", Type: "string"}}
	if !reflect.DeepEqual(c.Variables(), wantVars) {
		t.Errorf("variables: got %+v, want %+v", c.Variables(), wantVars)
	}

	// An auth mode the provider's definition does not list cannot be exported.
	if _, ok := c.Resources["metadata"]; ok {
		t.Error("want the connection with an unsupported auth_mode skipped")
	}

	// The data source reads name, so configuring it fails validation.
	var tf bytes.Buffer
	if err := c.GenerateTerraformFiles(context.Background(), &tf, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(tf.String(), `data "polytomic_apple_ads_attribution_connection" "attribution"`) {
		t.Fatalf("missing data source in:\n%s", tf.String())
	}
	if strings.Contains(tf.String(), "Attribution\"") {
		t.Errorf("data source configures name:\n%s", tf.String())
	}
}
