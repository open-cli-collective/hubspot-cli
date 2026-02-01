package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListBlogPosts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/blogs/posts", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "post-123",
						"name": "My First Post",
						"slug": "my-first-post",
						"state": "PUBLISHED",
						"authorName": "John Doe",
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

		result, err := client.ListBlogPosts(ListOptions{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "post-123", result.Results[0].ID)
		assert.Equal(t, "My First Post", result.Results[0].Name)
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

		result, err := client.ListBlogPosts(ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})
}

func TestClient_GetBlogPost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/blogs/posts/post-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "post-123",
				"name": "My First Post",
				"slug": "my-first-post",
				"state": "PUBLISHED",
				"authorName": "John Doe",
				"htmlTitle": "My First Blog Post",
				"metaDescription": "A great post",
				"postSummary": "This is a summary",
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

		post, err := client.GetBlogPost("post-123")
		require.NoError(t, err)
		assert.Equal(t, "post-123", post.ID)
		assert.Equal(t, "My First Post", post.Name)
		assert.Equal(t, "John Doe", post.AuthorName)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Post not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		post, err := client.GetBlogPost("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, post)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		post, err := client.GetBlogPost("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post ID is required")
		assert.Nil(t, post)
	})
}

func TestClient_CreateBlogPost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/blogs/posts", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"id": "post-456",
				"name": "New Post",
				"slug": "new-post",
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
			"name": "New Post",
			"slug": "new-post",
		}
		post, err := client.CreateBlogPost(data)
		require.NoError(t, err)
		assert.Equal(t, "post-456", post.ID)
		assert.Equal(t, "New Post", post.Name)
	})
}

func TestClient_UpdateBlogPost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/blogs/posts/post-123", r.URL.Path)
			assert.Equal(t, http.MethodPatch, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "post-123",
				"name": "Updated Post",
				"slug": "updated-post",
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
			"name": "Updated Post",
		}
		post, err := client.UpdateBlogPost("post-123", updates)
		require.NoError(t, err)
		assert.Equal(t, "post-123", post.ID)
		assert.Equal(t, "Updated Post", post.Name)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		post, err := client.UpdateBlogPost("", map[string]interface{}{"name": "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post ID is required")
		assert.Nil(t, post)
	})
}

func TestClient_DeleteBlogPost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/blogs/posts/post-123", r.URL.Path)
			assert.Equal(t, http.MethodDelete, r.Method)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		err := client.DeleteBlogPost("post-123")
		require.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		err := client.DeleteBlogPost("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post ID is required")
	})
}

func TestClient_ListBlogAuthors(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/blogs/authors", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "author-123",
						"name": "johndoe",
						"fullName": "John Doe",
						"email": "john@example.com",
						"displayName": "John D."
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

		result, err := client.ListBlogAuthors(ListOptions{})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "author-123", result.Results[0].ID)
		assert.Equal(t, "John Doe", result.Results[0].FullName)
	})
}

func TestClient_ListBlogTags(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/blogs/tags", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "tag-123",
						"name": "Technology",
						"slug": "technology",
						"language": "en"
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

		result, err := client.ListBlogTags(ListOptions{})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "tag-123", result.Results[0].ID)
		assert.Equal(t, "Technology", result.Results[0].Name)
	})
}
