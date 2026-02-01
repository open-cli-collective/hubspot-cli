package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListWorkflows(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/automation/v4/flows", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "workflow-123",
						"name": "Welcome Email",
						"type": "CONTACT_FLOW",
						"isEnabled": true,
						"objectTypeId": "0-1",
						"revisionId": "rev-456",
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

		result, err := client.ListWorkflows(ListOptions{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "workflow-123", result.Results[0].ID)
		assert.Equal(t, "Welcome Email", result.Results[0].Name)
		assert.Equal(t, "CONTACT_FLOW", result.Results[0].Type)
		assert.True(t, result.Results[0].Enabled)
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

		result, err := client.ListWorkflows(ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})
}

func TestClient_GetWorkflow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/automation/v4/flows/workflow-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "workflow-123",
				"name": "Welcome Email",
				"type": "CONTACT_FLOW",
				"isEnabled": true,
				"objectTypeId": "0-1",
				"revisionId": "rev-456",
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

		workflow, err := client.GetWorkflow("workflow-123")
		require.NoError(t, err)
		assert.Equal(t, "workflow-123", workflow.ID)
		assert.Equal(t, "Welcome Email", workflow.Name)
		assert.Equal(t, "CONTACT_FLOW", workflow.Type)
		assert.True(t, workflow.Enabled)
		assert.Equal(t, "0-1", workflow.ObjectTypeID)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Workflow not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		workflow, err := client.GetWorkflow("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, workflow)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		workflow, err := client.GetWorkflow("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workflow ID is required")
		assert.Nil(t, workflow)
	})
}
