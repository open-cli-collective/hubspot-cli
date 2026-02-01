package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListPages(t *testing.T) {
	t.Run("success site pages", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/pages/site-pages", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "page-123",
						"name": "About Us",
						"slug": "about-us",
						"state": "PUBLISHED",
						"created": "2024-01-15T10:00:00Z",
						"updated": "2024-01-16T12:00:00Z"
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

		result, err := client.ListPages(PageTypeSite, ListOptions{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "page-123", result.Results[0].ID)
		assert.Equal(t, "About Us", result.Results[0].Name)
		assert.Equal(t, "abc123", result.Paging.Next.After)
	})

	t.Run("success landing pages", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/pages/landing-pages", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"results": []}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		result, err := client.ListPages(PageTypeLanding, ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})
}

func TestClient_GetPage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/pages/site-pages/page-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "page-123",
				"name": "About Us",
				"slug": "about-us",
				"state": "PUBLISHED",
				"htmlTitle": "About Our Company",
				"metaDescription": "Learn about us",
				"created": "2024-01-15T10:00:00Z",
				"updated": "2024-01-16T12:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		page, err := client.GetPage(PageTypeSite, "page-123")
		require.NoError(t, err)
		assert.Equal(t, "page-123", page.ID)
		assert.Equal(t, "About Us", page.Name)
		assert.Equal(t, "About Our Company", page.HTMLTitle)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Page not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		page, err := client.GetPage(PageTypeSite, "nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, page)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		page, err := client.GetPage(PageTypeSite, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "page ID is required")
		assert.Nil(t, page)
	})
}

func TestClient_CreatePage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/pages/site-pages", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"id": "page-456",
				"name": "New Page",
				"slug": "new-page",
				"state": "DRAFT",
				"created": "2024-01-20T10:00:00Z",
				"updated": "2024-01-20T10:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		data := map[string]interface{}{
			"name": "New Page",
			"slug": "new-page",
		}
		page, err := client.CreatePage(PageTypeSite, data)
		require.NoError(t, err)
		assert.Equal(t, "page-456", page.ID)
		assert.Equal(t, "New Page", page.Name)
	})
}

func TestClient_UpdatePage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/pages/site-pages/page-123", r.URL.Path)
			assert.Equal(t, http.MethodPatch, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "page-123",
				"name": "Updated Page",
				"slug": "updated-page",
				"state": "DRAFT",
				"created": "2024-01-15T10:00:00Z",
				"updated": "2024-01-21T10:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		updates := map[string]interface{}{
			"name": "Updated Page",
		}
		page, err := client.UpdatePage(PageTypeSite, "page-123", updates)
		require.NoError(t, err)
		assert.Equal(t, "page-123", page.ID)
		assert.Equal(t, "Updated Page", page.Name)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		page, err := client.UpdatePage(PageTypeSite, "", map[string]interface{}{"name": "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "page ID is required")
		assert.Nil(t, page)
	})
}

func TestClient_DeletePage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/pages/site-pages/page-123", r.URL.Path)
			assert.Equal(t, http.MethodDelete, r.Method)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		err := client.DeletePage(PageTypeSite, "page-123")
		require.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		err := client.DeletePage(PageTypeSite, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "page ID is required")
	})
}

func TestClient_ClonePage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/pages/site-pages/page-123/clone", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"id": "page-789",
				"name": "About Us (Copy)",
				"slug": "about-us-copy",
				"state": "DRAFT",
				"created": "2024-01-21T10:00:00Z",
				"updated": "2024-01-21T10:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		page, err := client.ClonePage(PageTypeSite, "page-123")
		require.NoError(t, err)
		assert.Equal(t, "page-789", page.ID)
		assert.Equal(t, "About Us (Copy)", page.Name)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		page, err := client.ClonePage(PageTypeSite, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "page ID is required")
		assert.Nil(t, page)
	})
}
