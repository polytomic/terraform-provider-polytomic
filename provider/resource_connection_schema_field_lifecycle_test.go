package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ptclient "github.com/polytomic/polytomic-go/v25/client"
	"github.com/polytomic/polytomic-go/v25/option"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type schemaResponse struct {
	status int
	body   string
}

func schemaWithFields(fields string) schemaResponse {
	return schemaResponse{http.StatusOK, `{"data":{"id":"orders","fields":[` + fields + `]}}`}
}

var (
	schemaReadFailure  = schemaResponse{http.StatusInternalServerError, `{"message":"unavailable"}`}
	schemaNotFound     = schemaResponse{http.StatusNotFound, `{"message":"schema not found"}`}
	userManagedCity    = `{"id":"city","name":"City","type":"string","path":"$.address.city","user_managed":true}`
	detectedCity       = `{"id":"city","name":"city","type":"string"}`
	noFields           = ""
	testFieldSchemaURL = "/api/connections/conn-1/schemas/orders"
)

// schemaServer answers schema reads with responses in order, repeating the
// last, and accepts field changes. It returns the number of schema reads.
func schemaServer(t *testing.T, responses ...schemaResponse) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var reads atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == testFieldSchemaURL:
			resp := responses[min(int(reads.Add(1))-1, len(responses)-1)]
			w.WriteHeader(resp.status)
			fmt.Fprint(w, resp.body)
		case r.Method == http.MethodPost && r.URL.Path == testFieldSchemaURL+"/fields",
			r.Method == http.MethodDelete && r.URL.Path == testFieldSchemaURL+"/fields/city":
			w.WriteHeader(http.StatusAccepted)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"not found"}`)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &reads
}

func shortenFieldRemovalWait(t *testing.T) {
	timeout, interval := fieldRemovalTimeout, fieldRemovalInterval
	fieldRemovalTimeout, fieldRemovalInterval = 50*time.Millisecond, time.Millisecond
	t.Cleanup(func() { fieldRemovalTimeout, fieldRemovalInterval = timeout, interval })
}

func testFieldResource(t *testing.T, srv *httptest.Server) (*connectionSchemaFieldResource, resource.SchemaResponse) {
	t.Helper()
	p, err := providerclient.NewClientProvider(providerclient.Options{APIKey: "token", DeploymentURL: srv.URL})
	require.NoError(t, err)
	r := &connectionSchemaFieldResource{provider: p}
	var s resource.SchemaResponse
	r.Schema(t.Context(), resource.SchemaRequest{}, &s)
	require.False(t, s.Diagnostics.HasError(), "%v", s.Diagnostics)
	return r, s
}

func TestWaitForFieldRemoval(t *testing.T) {
	shortenFieldRemovalWait(t)

	for _, tc := range []struct {
		name      string
		responses []schemaResponse
		wantReads int32
		wantErr   bool
	}{
		{"added field leaves the schema", []schemaResponse{schemaWithFields(userManagedCity), schemaWithFields(userManagedCity), schemaWithFields(noFields)}, 3, false},
		{"overridden field reverts", []schemaResponse{schemaWithFields(userManagedCity), schemaWithFields(detectedCity)}, 2, false},
		{"schema is gone", []schemaResponse{schemaNotFound}, 1, false},
		{"a read fails", []schemaResponse{schemaReadFailure, schemaWithFields(noFields)}, 2, false},
		{"field stays user-defined", []schemaResponse{schemaWithFields(userManagedCity)}, 0, true},
		{"reads keep failing", []schemaResponse{schemaReadFailure}, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, reads := schemaServer(t, tc.responses...)
			client := ptclient.NewClient(option.WithBaseURL(srv.URL), option.WithToken("token"), option.WithMaxAttempts(1))

			err := waitForFieldRemoval(t.Context(), client, "conn-1", "orders", "city")
			if tc.wantErr != (err != nil) {
				t.Fatalf("got error %v, want error: %t", err, tc.wantErr)
			}
			if tc.wantReads != 0 {
				assert.Equal(t, tc.wantReads, reads.Load(), "schema reads")
			}
		})
	}
}

func TestConnectionSchemaFieldDelete(t *testing.T) {
	shortenFieldRemovalWait(t)

	for _, tc := range []struct {
		name        string
		reads       []schemaResponse
		wantWarning bool
	}{
		{"waits for the schema refresh", []schemaResponse{schemaWithFields(userManagedCity), schemaWithFields(noFields)}, false},
		{"warns when the refresh does not finish", []schemaResponse{schemaWithFields(userManagedCity)}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, reads := schemaServer(t, tc.reads...)
			r, s := testFieldResource(t, srv)
			ctx := t.Context()

			state := tfsdk.State{Schema: s.Schema}
			diags := state.Set(ctx, &connectionSchemaFieldResourceModel{
				ID:           types.StringValue("default/conn-1/orders/city"),
				Organization: types.StringValue(""),
				ConnectionID: types.StringValue("conn-1"),
				SchemaID:     types.StringValue("orders"),
				FieldID:      types.StringValue("city"),
				Label:        types.StringValue("City"),
				Type:         types.StringValue("string"),
				Precision:    types.Int64Null(),
				Scale:        types.Int64Null(),
				TypeSpec:     newTypeSpecNull(),
				Path:         types.StringValue("$.address.city"),
			})
			require.False(t, diags.HasError(), "%v", diags)

			resp := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.wantWarning, resp.Diagnostics.WarningsCount() > 0, "%v", resp.Diagnostics)
			assert.GreaterOrEqual(t, reads.Load(), int32(2), "schema reads")
		})
	}
}
