package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListFiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/files/v3/files", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "file-123",
						"name": "image.png",
						"path": "/images/image.png",
						"size": 12345,
						"type": "IMG",
						"extension": "png",
						"url": "https://example.com/image.png",
						"access": "PUBLIC_INDEXABLE",
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

		result, err := client.ListFiles(ListOptions{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "file-123", result.Results[0].ID)
		assert.Equal(t, "image.png", result.Results[0].Name)
		assert.Equal(t, int64(12345), result.Results[0].Size)
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

		result, err := client.ListFiles(ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})
}

func TestClient_GetFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/files/v3/files/file-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "file-123",
				"name": "document.pdf",
				"path": "/docs/document.pdf",
				"size": 54321,
				"type": "DOCUMENT",
				"extension": "pdf",
				"url": "https://example.com/document.pdf",
				"access": "PRIVATE",
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

		file, err := client.GetFile("file-123")
		require.NoError(t, err)
		assert.Equal(t, "file-123", file.ID)
		assert.Equal(t, "document.pdf", file.Name)
		assert.Equal(t, "PRIVATE", file.AccessLevel)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "File not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		file, err := client.GetFile("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, file)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		file, err := client.GetFile("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file ID is required")
		assert.Nil(t, file)
	})
}

func TestClient_DeleteFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/files/v3/files/file-123", r.URL.Path)
			assert.Equal(t, http.MethodDelete, r.Method)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		err := client.DeleteFile("file-123")
		require.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		err := client.DeleteFile("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file ID is required")
	})
}

func TestClient_ListFolders(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/files/v3/folders", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "folder-123",
						"name": "images",
						"path": "/images",
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

		result, err := client.ListFolders(ListOptions{})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "folder-123", result.Results[0].ID)
		assert.Equal(t, "images", result.Results[0].Name)
	})
}

func TestClient_ListDomains(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/domains", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "domain-123",
						"domain": "www.example.com",
						"primarySitePage": true,
						"isResolving": true,
						"isSslEnabled": true,
						"isSslOnly": true,
						"createdAt": "2024-01-15T10:00:00Z",
						"updatedAt": "2024-01-16T12:00:00Z"
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

		result, err := client.ListDomains(ListOptions{})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "domain-123", result.Results[0].ID)
		assert.Equal(t, "www.example.com", result.Results[0].Domain)
		assert.True(t, result.Results[0].PrimarySitePage)
		assert.True(t, result.Results[0].IsSslEnabled)
	})
}

func TestClient_GetDomain(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/domains/domain-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "domain-123",
				"domain": "blog.example.com",
				"primaryBlogPost": true,
				"isResolving": true,
				"isSslEnabled": true,
				"isSslOnly": false,
				"isUsedForBlogPost": true,
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

		domain, err := client.GetDomain("domain-123")
		require.NoError(t, err)
		assert.Equal(t, "domain-123", domain.ID)
		assert.Equal(t, "blog.example.com", domain.Domain)
		assert.True(t, domain.PrimaryBlogPost)
		assert.True(t, domain.IsUsedForBlogPost)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Domain not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		domain, err := client.GetDomain("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, domain)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		domain, err := client.GetDomain("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "domain ID is required")
		assert.Nil(t, domain)
	})
}
