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
	testOrgID          = "22c86135-fc64-4d26-8d32-c9c79079f070"
)

// schemaServer answers schema reads with responses in order, repeating the
// last, accepts field changes, and reports that conn-1 belongs to testOrgID.
// It returns the number of schema reads.
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
		case r.Method == http.MethodGet && r.URL.Path == "/api/connections/conn-1":
			fmt.Fprintf(w, `{"data":{"id":"conn-1","organization_id":%q}}`, testOrgID)
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

// cityPlan plans adding the city field without an organization.
func cityPlan(t *testing.T, s resource.SchemaResponse) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: s.Schema}
	diags := plan.Set(t.Context(), &connectionSchemaFieldResourceModel{
		ID:           types.StringUnknown(),
		Organization: types.StringUnknown(),
		ConnectionID: types.StringValue("conn-1"),
		SchemaID:     types.StringValue("orders"),
		FieldID:      types.StringValue("city"),
		Label:        types.StringValue("City"),
		Type:         types.StringValue("string"),
		Precision:    types.Int64Unknown(),
		Scale:        types.Int64Unknown(),
		TypeSpec:     newTypeSpecUnknown(),
		Path:         types.StringUnknown(),
	})
	require.False(t, diags.HasError(), "%v", diags)
	return plan
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

func TestConnectionSchemaFieldCreate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		readBack    schemaResponse
		wantWarning bool
		wantPath    types.String
	}{
		{"reads the added field back", schemaWithFields(userManagedCity), false, types.StringValue("$.address.city")},
		// The field exists once added, so it is recorded even though its
		// computed attributes are unknown.
		{"records the field when the read fails", schemaReadFailure, true, types.StringNull()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := schemaServer(t, schemaWithFields(noFields), tc.readBack)
			r, s := testFieldResource(t, srv)
			ctx := t.Context()

			resp := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
			r.Create(ctx, resource.CreateRequest{Plan: cityPlan(t, s)}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.wantWarning, resp.Diagnostics.WarningsCount() > 0, "%v", resp.Diagnostics)
			require.True(t, resp.State.Raw.IsFullyKnown(), "state has unknown values: %v", resp.State.Raw)

			var got connectionSchemaFieldResourceModel
			diags := resp.State.Get(ctx, &got)
			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, testOrgID, got.Organization.ValueString())
			assert.Equal(t, testOrgID+"/conn-1/orders/city", got.ID.ValueString())
			assert.Equal(t, "City", got.Label.ValueString())
			assert.Equal(t, "string", got.Type.ValueString())
			assert.Equal(t, tc.wantPath, got.Path)
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
				ID: types.StringValue("default/conn-1/orders/city"),
				// Recorded when Create could not determine the organization.
				Organization: types.StringValue(providerclient.DefaultOrganization),
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

func TestConnectionSchemaFieldCreateUserManaged(t *testing.T) {
	srv, _ := schemaServer(t, schemaWithFields(userManagedCity))
	r, s := testFieldResource(t, srv)

	resp := resource.CreateResponse{State: tfsdk.State{Schema: s.Schema}}
	r.Create(t.Context(), resource.CreateRequest{Plan: cityPlan(t, s)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	// The suggested import ID names the connection's organization.
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), testOrgID+"/conn-1/orders/city")
}

func TestResourceOrganization(t *testing.T) {
	srv, _ := schemaServer(t, schemaWithFields(noFields))
	client := ptclient.NewClient(option.WithBaseURL(srv.URL), option.WithToken("token"), option.WithMaxAttempts(1))

	for _, tc := range []struct {
		name         string
		org          types.String
		connectionID string
		want         string
	}{
		{"configured", types.StringValue("org-1"), "conn-1", "org-1"},
		{"connection's", types.StringUnknown(), "conn-1", testOrgID},
		{"lookup fails", types.StringNull(), "conn-2", providerclient.DefaultOrganization},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, resourceOrganization(t.Context(), client, tc.org, tc.connectionID))
		})
	}
}
