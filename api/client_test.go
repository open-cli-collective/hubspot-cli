package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		cfg         ClientConfig
		wantErr     error
		wantBaseURL string
	}{
		{
			name: "valid config",
			cfg: ClientConfig{
				AccessToken: "pat-na1-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			},
			wantErr:     nil,
			wantBaseURL: DefaultBaseURL,
		},
		{
			name: "valid config with verbose",
			cfg: ClientConfig{
				AccessToken: "pat-eu1-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				Verbose:     true,
			},
			wantErr:     nil,
			wantBaseURL: DefaultBaseURL,
		},
		{
			name:    "missing access token",
			cfg:     ClientConfig{},
			wantErr: ErrAccessTokenRequired,
		},
		{
			name: "empty access token",
			cfg: ClientConfig{
				AccessToken: "",
			},
			wantErr: ErrAccessTokenRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.cfg)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.wantBaseURL, client.BaseURL)
				assert.Equal(t, tt.cfg.AccessToken, client.AccessToken)
				assert.Equal(t, tt.cfg.Verbose, client.Verbose)
				assert.NotNil(t, client.HTTPClient)
			}
		})
	}
}

func TestClient_authHeader(t *testing.T) {
	client := &Client{
		AccessToken: "pat-na1-test-token-value",
	}

	header := client.authHeader()

	assert.Equal(t, "Bearer pat-na1-test-token-value", header)
}

func TestClient_doRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		responseStatus int
		responseBody   string
		wantErr        bool
	}{
		{
			name:           "successful GET",
			method:         http.MethodGet,
			responseStatus: http.StatusOK,
			responseBody:   `{"key": "value"}`,
			wantErr:        false,
		},
		{
			name:           "successful POST",
			method:         http.MethodPost,
			responseStatus: http.StatusCreated,
			responseBody:   `{"id": "123"}`,
			wantErr:        false,
		},
		{
			name:           "successful PATCH",
			method:         http.MethodPatch,
			responseStatus: http.StatusOK,
			responseBody:   `{"updated": true}`,
			wantErr:        false,
		},
		{
			name:           "successful DELETE",
			method:         http.MethodDelete,
			responseStatus: http.StatusNoContent,
			responseBody:   ``,
			wantErr:        false,
		},
		{
			name:           "unauthorized",
			method:         http.MethodGet,
			responseStatus: http.StatusUnauthorized,
			responseBody:   `{"status": "error", "message": "Invalid access token"}`,
			wantErr:        true,
		},
		{
			name:           "not found",
			method:         http.MethodGet,
			responseStatus: http.StatusNotFound,
			responseBody:   `{"status": "error", "message": "Contact not found"}`,
			wantErr:        true,
		},
		{
			name:           "server error",
			method:         http.MethodGet,
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"status": "error", "message": "Internal error"}`,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify auth header is Bearer
				authHeader := r.Header.Get("Authorization")
				assert.True(t, len(authHeader) > 7 && authHeader[:7] == "Bearer ")
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				AccessToken: "test-token",
				HTTPClient:  server.Client(),
			}

			body, err := client.doRequest(tt.method, server.URL, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.responseBody, string(body))
			}
		})
	}
}

func TestClient_doRequest_withBody(t *testing.T) {
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedBody)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := &Client{
		AccessToken: "test-token",
		HTTPClient:  server.Client(),
	}

	requestBody := map[string]interface{}{
		"email":     "john@example.com",
		"firstname": "John",
		"lastname":  "Doe",
	}

	_, err := client.doRequest(http.MethodPost, server.URL, requestBody)
	require.NoError(t, err)

	assert.Equal(t, "john@example.com", receivedBody["email"])
	assert.Equal(t, "John", receivedBody["firstname"])
	assert.Equal(t, "Doe", receivedBody["lastname"])
}

func TestClient_httpMethods(t *testing.T) {
	// Track which methods were called
	var calledMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledMethod = r.Method
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := &Client{
		AccessToken: "test-token",
		HTTPClient:  server.Client(),
	}

	t.Run("get", func(t *testing.T) {
		_, err := client.get(server.URL)
		require.NoError(t, err)
		assert.Equal(t, http.MethodGet, calledMethod)
	})

	t.Run("post", func(t *testing.T) {
		_, err := client.post(server.URL, nil)
		require.NoError(t, err)
		assert.Equal(t, http.MethodPost, calledMethod)
	})

	t.Run("patch", func(t *testing.T) {
		_, err := client.patch(server.URL, nil)
		require.NoError(t, err)
		assert.Equal(t, http.MethodPatch, calledMethod)
	})

	t.Run("delete", func(t *testing.T) {
		_, err := client.delete(server.URL)
		require.NoError(t, err)
		assert.Equal(t, http.MethodDelete, calledMethod)
	})
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		params map[string]string
		want   string
	}{
		{
			name:   "no params",
			base:   "https://api.hubapi.com/crm/v3/objects/contacts",
			params: nil,
			want:   "https://api.hubapi.com/crm/v3/objects/contacts",
		},
		{
			name:   "empty params",
			base:   "https://api.hubapi.com/crm/v3/objects/contacts",
			params: map[string]string{},
			want:   "https://api.hubapi.com/crm/v3/objects/contacts",
		},
		{
			name: "single param",
			base: "https://api.hubapi.com/crm/v3/objects/contacts",
			params: map[string]string{
				"limit": "10",
			},
			want: "https://api.hubapi.com/crm/v3/objects/contacts?limit=10",
		},
		{
			name: "multiple params",
			base: "https://api.hubapi.com/crm/v3/objects/contacts",
			params: map[string]string{
				"limit": "10",
				"after": "abc123",
			},
			want: "https://api.hubapi.com/crm/v3/objects/contacts?after=abc123&limit=10",
		},
		{
			name: "skip empty values",
			base: "https://api.hubapi.com/crm/v3/objects/contacts",
			params: map[string]string{
				"limit":      "10",
				"properties": "",
			},
			want: "https://api.hubapi.com/crm/v3/objects/contacts?limit=10",
		},
		{
			name: "encode special characters",
			base: "https://api.hubapi.com/crm/v3/objects/contacts/search",
			params: map[string]string{
				"q": "email = john@example.com",
			},
			want: "https://api.hubapi.com/crm/v3/objects/contacts/search?q=email+%3D+john%40example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildURL(tt.base, tt.params)
			assert.Equal(t, tt.want, got)
		})
	}
}
