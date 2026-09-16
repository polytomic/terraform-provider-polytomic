package importer

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	ptclient "github.com/polytomic/polytomic-go/client"
	"github.com/polytomic/polytomic-go/option"
)

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

// TestSyncsFieldOverrideValue verifies that a field mapping with an
// override_value is exported, alongside the sync's override fields.
func TestSyncsFieldOverrideValue(t *testing.T) {
	s := NewSyncs(newTestClient(t, map[string]string{
		"/api/syncs": `{"data":[{"id":"sync-1","name":"Overrides"}]}`,
		"/api/syncs/sync-1": `{"data":{"id":"sync-1","name":"Overrides","active":false,"mode":"replace",
			"schedule":{"frequency":"manual"},
			"target":{"connection_id":"conn-1","object":"contacts"},
			"fields":[
				{"source":{"model_id":"model-1","field":"email"},"target":"email"},
				{"source":{"model_id":"model-1","field":"email"},"target":"name","override_value":"synced"}
			],
			"override_fields":[{"target":"status","override_value":"active"}]}}`,
	}))
	if err := s.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	var tf bytes.Buffer
	if err := s.GenerateTerraformFiles(context.Background(), &tf, nil); err != nil {
		t.Fatal(err)
	}
	if _, diags := hclsyntax.ParseConfig(tf.Bytes(), SyncResourceFileName, hcl.InitialPos); diags.HasErrors() {
		t.Fatalf("%s\n%s", diags, tf.String())
	}
	for _, want := range []string{
		`target\s*=\s*"email"`,
		`override_value\s*=\s*"synced"`,
		`target\s*=\s*"name"`,
		`override_fields\s*=`,
		`override_value\s*=\s*"active"`,
	} {
		if !regexp.MustCompile(want).Match(tf.Bytes()) {
			t.Errorf("missing %s in:\n%s", want, tf.String())
		}
	}
}
