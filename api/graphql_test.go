package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ExecuteGraphQL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Basic assertions that don't involve reading body
			if r.URL.Path != "/collector/graphql" {
				t.Errorf("expected path /collector/graphql, got %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"CRM": {
						"contact_collection": {
							"items": [
								{"firstname": "John", "lastname": "Doe", "email": "john@example.com"}
							],
							"total": 1
						}
					}
				}
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		query := `query { CRM { contact_collection(limit: 10) { items { firstname lastname email } total } } }`
		result, err := client.ExecuteGraphQL(query, nil)
		require.NoError(t, err)
		assert.False(t, result.HasErrors())
		assert.NotNil(t, result.Data)
	})

	t.Run("with variables", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"CRM": {
						"contact": {"firstname": "Jane", "email": "jane@example.com"}
					}
				}
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		query := `query GetContact($contactId: ID!) { CRM { contact(id: $contactId) { firstname email } } }`
		vars := map[string]interface{}{"contactId": "123"}
		result, err := client.ExecuteGraphQL(query, vars)
		require.NoError(t, err)
		assert.False(t, result.HasErrors())
	})

	t.Run("graphql error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": null,
				"errors": [
					{
						"message": "Cannot query field 'invalid_field' on type 'Contact'",
						"locations": [{"line": 1, "column": 15}],
						"path": ["CRM", "contact"]
					}
				]
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		query := `query { CRM { contact(id: "123") { invalid_field } } }`
		result, err := client.ExecuteGraphQL(query, nil)
		require.NoError(t, err)
		assert.True(t, result.HasErrors())
		assert.Len(t, result.Errors, 1)
		assert.Contains(t, result.Errors[0].Message, "invalid_field")
	})

	t.Run("multiple errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": null,
				"errors": [
					{"message": "Error 1"},
					{"message": "Error 2"}
				]
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		result, err := client.ExecuteGraphQL("query { invalid }", nil)
		require.NoError(t, err)
		assert.True(t, result.HasErrors())
		assert.Equal(t, "Error 1; Error 2", result.ErrorMessages())
	})

	t.Run("empty query", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		result, err := client.ExecuteGraphQL("", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query is required")
		assert.Nil(t, result)
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status": "error", "message": "Invalid token"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "bad-token",
			HTTPClient:  server.Client(),
		}

		result, err := client.ExecuteGraphQL("query { CRM { contact(id: \"123\") { email } } }", nil)
		assert.Error(t, err)
		assert.True(t, IsUnauthorized(err))
		assert.Nil(t, result)
	})
}

func TestGraphQLResponse_HasErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		resp := &GraphQLResponse{
			Data: json.RawMessage(`{"test": "data"}`),
		}
		assert.False(t, resp.HasErrors())
	})

	t.Run("with errors", func(t *testing.T) {
		resp := &GraphQLResponse{
			Errors: []GraphQLError{{Message: "test error"}},
		}
		assert.True(t, resp.HasErrors())
	})
}

func TestGraphQLResponse_ErrorMessages(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		resp := &GraphQLResponse{}
		assert.Equal(t, "", resp.ErrorMessages())
	})

	t.Run("single error", func(t *testing.T) {
		resp := &GraphQLResponse{
			Errors: []GraphQLError{{Message: "Single error"}},
		}
		assert.Equal(t, "Single error", resp.ErrorMessages())
	})

	t.Run("multiple errors", func(t *testing.T) {
		resp := &GraphQLResponse{
			Errors: []GraphQLError{
				{Message: "First"},
				{Message: "Second"},
				{Message: "Third"},
			},
		}
		assert.Equal(t, "First; Second; Third", resp.ErrorMessages())
	})
}
