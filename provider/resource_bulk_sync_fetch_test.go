package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	ptclient "github.com/polytomic/polytomic-go/v25/client"
	"github.com/polytomic/polytomic-go/v25/option"
	"github.com/stretchr/testify/require"
)

func TestFetchBulkSyncSchemasSelection(t *testing.T) {
	for _, tc := range []struct {
		name       string
		ids        []string
		wantFilter bool
		wantID     string
	}{
		{name: "import enabled selections", wantFilter: true, wantID: "selected"},
		{name: "explicit disabled schema", ids: []string{"unselected"}, wantID: "unselected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/api/bulk/syncs/sync-1/schemas":
					query, _ := url.QueryUnescape(r.URL.RawQuery)
					filtered := strings.Contains(query, "enabled") && strings.Contains(query, "true")
					if filtered != tc.wantFilter {
						t.Errorf("unexpected schema selection filter: %s", query)
					}
					if filtered {
						fmt.Fprint(w, `{"data":[{"id":"selected","enabled":true}]}`)
					} else {
						fmt.Fprint(w, `{"data":[{"id":"selected","enabled":true},{"id":"unselected","enabled":false}]}`)
					}
				case "/api/bulk/syncs/sync-1/schemas/selected":
					fmt.Fprint(w, `{"data":{"id":"selected","enabled":true}}`)
				case "/api/bulk/syncs/sync-1/schemas/unselected":
					fmt.Fprint(w, `{"data":{"id":"unselected","enabled":false}}`)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			client := ptclient.NewClient(option.WithBaseURL(server.URL), option.WithToken("test-token"))
			schemas, err := fetchBulkSyncSchemas(t.Context(), client, "sync-1", tc.ids)
			require.NoError(t, err)
			require.Len(t, schemas, 1)
			require.Equal(t, tc.wantID, *schemas[0].ID)
		})
	}
}
