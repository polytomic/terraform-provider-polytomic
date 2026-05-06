package providerclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientProvider_URLNormalization(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		expectedURL string
	}{
		{
			name:        "empty URL uses default",
			inputURL:    "",
			expectedURL: "https://app.polytomic.com",
		},
		{
			name:        "URL without scheme gets https prefix",
			inputURL:    "app.polytomic.com",
			expectedURL: "https://app.polytomic.com",
		},
		{
			name:        "URL with https scheme is preserved",
			inputURL:    "https://app.polytomic.com",
			expectedURL: "https://app.polytomic.com",
		},
		{
			name:        "URL with http scheme is preserved",
			inputURL:    "http://localhost:8080",
			expectedURL: "http://localhost:8080",
		},
		{
			name:        "trailing slash is removed",
			inputURL:    "https://app.polytomic.com/",
			expectedURL: "https://app.polytomic.com",
		},
		{
			name:        "multiple trailing slashes are removed",
			inputURL:    "https://app.polytomic.com///",
			expectedURL: "https://app.polytomic.com",
		},
		{
			name:        "custom domain without scheme",
			inputURL:    "custom.polytomic-local.com",
			expectedURL: "https://custom.polytomic-local.com",
		},
		{
			name:        "localhost with port without scheme",
			inputURL:    "localhost:8080",
			expectedURL: "https://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewClientProvider(Options{
				DeploymentURL: tt.inputURL,
				APIKey:        "test-key", // Required to pass validation
			})
			require.NoError(t, err)
			assert.Equal(t, tt.expectedURL, provider.opts.DeploymentURL)
		})
	}
}

// TestClient_APIKeyOrgVerificationUsesIdentity pins the regression where the
// API-key path verified the requested organization by calling
// /api/organizations. That endpoint requires a partner key on current API
// versions, so the call fails and the provider used to surface a misleading
// "partner key is required" error.
//
// The fix is to verify the org via /api/me (Identity.Get), which works with
// API keys. The fake server here returns 403 for /api/organizations and a
// matching identity for /api/me; Client(ctx, org) must succeed without ever
// calling /api/organizations.
func TestClient_APIKeyOrgVerificationUsesIdentity(t *testing.T) {
	orgID := uuid.NewString()

	var organizationsHits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/me":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":              "user-1",
					"organization_id": orgID,
				},
			})
		case "/api/organizations":
			atomic.AddInt32(&organizationsHits, 1)
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"partner key required"}`))
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	provider, err := NewClientProvider(Options{
		APIKey:        "test-api-key",
		DeploymentURL: server.URL,
	})
	require.NoError(t, err)

	c, err := provider.Client(context.Background(), orgID)
	require.NoError(t, err, "Client should succeed for an API key whose org matches the request")
	assert.NotNil(t, c)
	assert.Zero(t, atomic.LoadInt32(&organizationsHits),
		"Client must not hit /api/organizations on the API-key path")
}

// TestClient_APIKeyOrgMismatchRejected verifies the negative case: when the
// requested organization does not match what /api/me reports for the API key,
// Client(ctx, org) returns an error rather than silently handing back a
// misconfigured client.
func TestClient_APIKeyOrgMismatchRejected(t *testing.T) {
	keyOrgID := uuid.NewString()
	requestedOrgID := uuid.NewString()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/me" {
			t.Errorf("unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":              "user-1",
				"organization_id": keyOrgID,
			},
		})
	}))
	defer server.Close()

	provider, err := NewClientProvider(Options{
		APIKey:        "test-api-key",
		DeploymentURL: server.URL,
	})
	require.NoError(t, err)

	_, err = provider.Client(context.Background(), requestedOrgID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), requestedOrgID,
		"error should name the rejected organization id")
}

func TestNewClientProvider_InvalidURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
	}{
		{
			name:     "invalid URL with spaces",
			inputURL: "https://app polytomic.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientProvider(Options{
				DeploymentURL: tt.inputURL,
				APIKey:        "test-key",
			})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid deployment URL")
		})
	}
}
