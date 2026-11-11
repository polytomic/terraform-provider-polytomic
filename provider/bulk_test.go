package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
	"github.com/stretchr/testify/require"
)

func TestBulkDataSourcesCapabilities(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/connections/source/bulksync/source":
			fmt.Fprint(w, `{"data":{"schemas":[],"capabilities":{"supports_selective_modes":true}}}`)
		case "/api/connections/destination/bulksync/target":
			if r.URL.Query().Get("source_connection_id") != "source" {
				t.Errorf("missing source connection query: %s", r.URL.RawQuery)
			}
			fmt.Fprint(w, `{"data":{"configuration":{"schema":{}},"initial_configuration":{"schema":"suggested"},"supports_history_mode":true,"modes":[{"id":"replicate","label":"Replicate","description":"Replicate tables"}]}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()
	p, err := providerclient.NewClientProvider(providerclient.Options{APIKey: "test", DeploymentURL: api.URL})
	require.NoError(t, err)
	for _, tc := range []struct {
		name   string
		source datasource.DataSource
		config string
		check  func(map[string]tftypes.Value)
	}{
		{"source", &bulkSourceDatasource{provider: p}, `{"connection_id":"source"}`, func(values map[string]tftypes.Value) {
			var supported bool
			require.NoError(t, values["supports_selective_modes"].As(&supported))
			require.True(t, supported)
		}},
		{"destination", &bulkDestinationDatasource{provider: p}, `{"connection_id":"destination","source_connection_id":"source"}`, func(values map[string]tftypes.Value) {
			var supported bool
			require.NoError(t, values["supports_history_mode"].As(&supported))
			require.True(t, supported)
			var initial string
			require.NoError(t, values["initial_configuration"].As(&initial))
			require.JSONEq(t, `{"schema":"suggested"}`, initial)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var schema datasource.SchemaResponse
			tc.source.Schema(t.Context(), datasource.SchemaRequest{}, &schema)
			raw, err := tftypes.ValueFromJSON([]byte(tc.config), schema.Schema.Type().TerraformType(t.Context()))
			require.NoError(t, err)
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
			tc.source.Read(t.Context(), datasource.ReadRequest{Config: tfsdk.Config{Raw: raw, Schema: schema.Schema}}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var values map[string]tftypes.Value
			require.NoError(t, resp.State.Raw.As(&values))
			tc.check(values)
		})
	}
}
