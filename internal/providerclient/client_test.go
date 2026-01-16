package providerclient

import (
	"testing"

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
