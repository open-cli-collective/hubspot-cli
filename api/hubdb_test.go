package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListHubDBTables(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "123",
						"name": "products",
						"label": "Products Table",
						"published": true,
						"rowCount": 50,
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

		result, err := client.ListHubDBTables(ListOptions{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "123", result.Results[0].ID)
		assert.Equal(t, "products", result.Results[0].Name)
		assert.Equal(t, 50, result.Results[0].RowCount)
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

		result, err := client.ListHubDBTables(ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})
}

func TestClient_GetHubDBTable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables/products", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "123",
				"name": "products",
				"label": "Products Table",
				"published": true,
				"rowCount": 50,
				"columns": [
					{
						"id": "1",
						"name": "name",
						"label": "Product Name",
						"type": "TEXT"
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

		table, err := client.GetHubDBTable("products")
		require.NoError(t, err)
		assert.Equal(t, "123", table.ID)
		assert.Equal(t, "products", table.Name)
		assert.Len(t, table.Columns, 1)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Table not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		table, err := client.GetHubDBTable("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, table)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		table, err := client.GetHubDBTable("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table ID or name is required")
		assert.Nil(t, table)
	})
}

func TestClient_CreateHubDBTable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"id": "456",
				"name": "new_table",
				"label": "New Table",
				"published": false
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		data := map[string]interface{}{
			"name":  "new_table",
			"label": "New Table",
		}
		table, err := client.CreateHubDBTable(data)
		require.NoError(t, err)
		assert.Equal(t, "456", table.ID)
		assert.Equal(t, "new_table", table.Name)
	})
}

func TestClient_DeleteHubDBTable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables/products", r.URL.Path)
			assert.Equal(t, http.MethodDelete, r.Method)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		err := client.DeleteHubDBTable("products")
		require.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		err := client.DeleteHubDBTable("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table ID or name is required")
	})
}

func TestClient_PublishHubDBTable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables/products/draft/publish", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "123",
				"name": "products",
				"published": true,
				"publishedAt": "2024-01-20T10:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		table, err := client.PublishHubDBTable("products")
		require.NoError(t, err)
		assert.Equal(t, "123", table.ID)
		assert.True(t, table.Published)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		table, err := client.PublishHubDBTable("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table ID or name is required")
		assert.Nil(t, table)
	})
}

func TestClient_ListHubDBRows(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables/products/rows", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "row-1",
						"path": "/products/widget",
						"name": "Widget",
						"values": {
							"1": "Widget Name",
							"2": 29.99
						}
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

		result, err := client.ListHubDBRows("products", ListOptions{})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "row-1", result.Results[0].ID)
		assert.Equal(t, "Widget", result.Results[0].Name)
	})

	t.Run("empty table ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		result, err := client.ListHubDBRows("", ListOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table ID or name is required")
		assert.Nil(t, result)
	})
}

func TestClient_GetHubDBRow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables/products/rows/row-1", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "row-1",
				"path": "/products/widget",
				"name": "Widget",
				"values": {
					"1": "Widget Name",
					"2": 29.99
				}
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		row, err := client.GetHubDBRow("products", "row-1")
		require.NoError(t, err)
		assert.Equal(t, "row-1", row.ID)
		assert.Equal(t, "Widget", row.Name)
	})

	t.Run("empty table ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		row, err := client.GetHubDBRow("", "row-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table ID or name is required")
		assert.Nil(t, row)
	})

	t.Run("empty row ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		row, err := client.GetHubDBRow("products", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "row ID is required")
		assert.Nil(t, row)
	})
}

func TestClient_CreateHubDBRow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables/products/rows/draft", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"id": "row-2",
				"path": "/products/gadget",
				"name": "Gadget"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		data := map[string]interface{}{
			"path": "/products/gadget",
			"name": "Gadget",
		}
		row, err := client.CreateHubDBRow("products", data)
		require.NoError(t, err)
		assert.Equal(t, "row-2", row.ID)
	})

	t.Run("empty table ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		row, err := client.CreateHubDBRow("", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table ID or name is required")
		assert.Nil(t, row)
	})
}

func TestClient_UpdateHubDBRow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables/products/rows/draft/row-1", r.URL.Path)
			assert.Equal(t, http.MethodPatch, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "row-1",
				"path": "/products/updated-widget",
				"name": "Updated Widget"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		updates := map[string]interface{}{
			"name": "Updated Widget",
		}
		row, err := client.UpdateHubDBRow("products", "row-1", updates)
		require.NoError(t, err)
		assert.Equal(t, "row-1", row.ID)
		assert.Equal(t, "Updated Widget", row.Name)
	})

	t.Run("empty table ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		row, err := client.UpdateHubDBRow("", "row-1", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table ID or name is required")
		assert.Nil(t, row)
	})

	t.Run("empty row ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		row, err := client.UpdateHubDBRow("products", "", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "row ID is required")
		assert.Nil(t, row)
	})
}

func TestClient_DeleteHubDBRow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/cms/v3/hubdb/tables/products/rows/draft/row-1", r.URL.Path)
			assert.Equal(t, http.MethodDelete, r.Method)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		err := client.DeleteHubDBRow("products", "row-1")
		require.NoError(t, err)
	})

	t.Run("empty table ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		err := client.DeleteHubDBRow("", "row-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table ID or name is required")
	})

	t.Run("empty row ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		err := client.DeleteHubDBRow("products", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "row ID is required")
	})
}
