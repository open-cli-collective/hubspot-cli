package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListSchemas(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/schemas", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "123",
						"name": "my_custom_object",
						"fullyQualifiedName": "p_my_custom_object",
						"labels": {
							"singular": "My Object",
							"plural": "My Objects"
						},
						"properties": [
							{
								"name": "name",
								"label": "Name",
								"type": "string"
							}
						],
						"createdAt": "2024-01-15T10:00:00Z",
						"updatedAt": "2024-01-16T12:00:00Z"
					}
				],
				"paging": {
					"next": {
						"after": "abc123"
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

		result, err := client.ListSchemas(ListOptions{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "123", result.Results[0].ID)
		assert.Equal(t, "my_custom_object", result.Results[0].Name)
		assert.Equal(t, "p_my_custom_object", result.Results[0].FullyQualifiedName)
		assert.Equal(t, "My Object", result.Results[0].Labels.Singular)
		assert.Equal(t, "abc123", result.Paging.Next.After)
	})

	t.Run("empty results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"results": []}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		result, err := client.ListSchemas(ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})
}

func TestClient_GetSchema(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/schemas/p_my_custom_object", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "123",
				"name": "my_custom_object",
				"fullyQualifiedName": "p_my_custom_object",
				"objectTypeId": "2-123456",
				"labels": {
					"singular": "My Object",
					"plural": "My Objects"
				},
				"primaryDisplayProperty": "name",
				"properties": [
					{
						"name": "name",
						"label": "Name",
						"type": "string",
						"fieldType": "text"
					}
				],
				"associatedObjects": ["contacts", "companies"],
				"createdAt": "2024-01-15T10:00:00Z",
				"updatedAt": "2024-01-16T12:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		schema, err := client.GetSchema("p_my_custom_object")
		require.NoError(t, err)
		assert.Equal(t, "123", schema.ID)
		assert.Equal(t, "my_custom_object", schema.Name)
		assert.Equal(t, "2-123456", schema.ObjectTypeID)
		assert.Len(t, schema.Properties, 1)
		assert.Len(t, schema.AssociatedObjects, 2)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Schema not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		schema, err := client.GetSchema("p_nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, schema)
	})

	t.Run("empty name", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		schema, err := client.GetSchema("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fully qualified name is required")
		assert.Nil(t, schema)
	})
}

func TestClient_CreateSchema(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/schemas", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"id": "456",
				"name": "new_object",
				"fullyQualifiedName": "p_new_object",
				"labels": {
					"singular": "New Object",
					"plural": "New Objects"
				},
				"createdAt": "2024-01-20T10:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		data := map[string]interface{}{
			"name": "new_object",
			"labels": map[string]string{
				"singular": "New Object",
				"plural":   "New Objects",
			},
		}
		schema, err := client.CreateSchema(data)
		require.NoError(t, err)
		assert.Equal(t, "456", schema.ID)
		assert.Equal(t, "new_object", schema.Name)
		assert.Equal(t, "p_new_object", schema.FullyQualifiedName)
	})
}

func TestClient_DeleteSchema(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/schemas/p_my_custom_object", r.URL.Path)
			assert.Equal(t, http.MethodDelete, r.Method)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		err := client.DeleteSchema("p_my_custom_object")
		require.NoError(t, err)
	})

	t.Run("empty name", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		err := client.DeleteSchema("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fully qualified name is required")
	})
}
