package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListInboxes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/conversations/v3/conversations/inboxes", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "inbox-123",
						"name": "Support",
						"createdAt": "2024-01-15T10:00:00Z",
						"updatedAt": "2024-01-16T12:00:00Z",
						"archived": false
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

		result, err := client.ListInboxes(ListOptions{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "inbox-123", result.Results[0].ID)
		assert.Equal(t, "Support", result.Results[0].Name)
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

		result, err := client.ListInboxes(ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})
}

func TestClient_GetInbox(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/conversations/v3/conversations/inboxes/inbox-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "inbox-123",
				"name": "Support",
				"createdAt": "2024-01-15T10:00:00Z",
				"updatedAt": "2024-01-16T12:00:00Z",
				"archived": false
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		inbox, err := client.GetInbox("inbox-123")
		require.NoError(t, err)
		assert.Equal(t, "inbox-123", inbox.ID)
		assert.Equal(t, "Support", inbox.Name)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Inbox not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		inbox, err := client.GetInbox("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, inbox)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		inbox, err := client.GetInbox("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inbox ID is required")
		assert.Nil(t, inbox)
	})
}

func TestClient_ListThreads(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/conversations/v3/conversations/threads", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "thread-123",
						"status": "OPEN",
						"associatedContactId": "contact-456",
						"inboxId": "inbox-789",
						"createdAt": "2024-01-15T10:00:00Z",
						"updatedAt": "2024-01-16T12:00:00Z",
						"archived": false
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

		result, err := client.ListThreads(ListOptions{})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "thread-123", result.Results[0].ID)
		assert.Equal(t, "OPEN", result.Results[0].Status)
		assert.Equal(t, "contact-456", result.Results[0].AssociatedContactID)
	})
}

func TestClient_GetThread(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/conversations/v3/conversations/threads/thread-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "thread-123",
				"status": "CLOSED",
				"associatedContactId": "contact-456",
				"inboxId": "inbox-789",
				"createdAt": "2024-01-15T10:00:00Z",
				"updatedAt": "2024-01-16T12:00:00Z",
				"closedAt": "2024-01-16T14:00:00Z",
				"archived": false
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		thread, err := client.GetThread("thread-123")
		require.NoError(t, err)
		assert.Equal(t, "thread-123", thread.ID)
		assert.Equal(t, "CLOSED", thread.Status)
		assert.Equal(t, "2024-01-16T14:00:00Z", thread.ClosedAt)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Thread not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		thread, err := client.GetThread("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, thread)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		thread, err := client.GetThread("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "thread ID is required")
		assert.Nil(t, thread)
	})
}
