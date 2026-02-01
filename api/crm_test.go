package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRMObject_GetProperty(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]interface{}
		propName   string
		want       string
	}{
		{
			name: "string property",
			properties: map[string]interface{}{
				"email": "john@example.com",
			},
			propName: "email",
			want:     "john@example.com",
		},
		{
			name: "number property",
			properties: map[string]interface{}{
				"amount": float64(1234.56),
			},
			propName: "amount",
			want:     "1234.56",
		},
		{
			name: "integer as float",
			properties: map[string]interface{}{
				"count": float64(42),
			},
			propName: "count",
			want:     "42",
		},
		{
			name: "bool property",
			properties: map[string]interface{}{
				"active": true,
			},
			propName: "active",
			want:     "true",
		},
		{
			name: "nil property",
			properties: map[string]interface{}{
				"email": nil,
			},
			propName: "email",
			want:     "",
		},
		{
			name: "missing property",
			properties: map[string]interface{}{
				"email": "test@example.com",
			},
			propName: "phone",
			want:     "",
		},
		{
			name:       "nil properties map",
			properties: nil,
			propName:   "email",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &CRMObject{Properties: tt.properties}
			got := obj.GetProperty(tt.propName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_ListObjects(t *testing.T) {
	t.Run("list contacts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/objects/contacts", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, "email,firstname,lastname", r.URL.Query().Get("properties"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "123",
						"properties": {
							"email": "john@example.com",
							"firstname": "John",
							"lastname": "Doe"
						},
						"createdAt": "2024-01-15T10:00:00Z",
						"updatedAt": "2024-01-15T10:00:00Z"
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

		result, err := client.ListObjects(ObjectTypeContacts, ListOptions{
			Limit:      10,
			Properties: []string{"email", "firstname", "lastname"},
		})

		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "123", result.Results[0].ID)
		assert.Equal(t, "john@example.com", result.Results[0].GetProperty("email"))
		assert.Equal(t, "abc123", result.Paging.Next.After)
	})

	t.Run("with pagination cursor", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "cursor123", r.URL.Query().Get("after"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"results": []}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		_, err := client.ListObjects(ObjectTypeContacts, ListOptions{After: "cursor123"})
		require.NoError(t, err)
	})
}

func TestClient_GetObject(t *testing.T) {
	t.Run("get contact by ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/objects/contacts/12345", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "12345",
				"properties": {
					"email": "jane@example.com",
					"firstname": "Jane",
					"lastname": "Smith"
				},
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

		obj, err := client.GetObject(ObjectTypeContacts, "12345", nil)
		require.NoError(t, err)
		assert.Equal(t, "12345", obj.ID)
		assert.Equal(t, "jane@example.com", obj.GetProperty("email"))
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Contact not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		obj, err := client.GetObject(ObjectTypeContacts, "99999", nil)
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, obj)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		obj, err := client.GetObject(ObjectTypeContacts, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "object ID is required")
		assert.Nil(t, obj)
	})
}

func TestClient_CreateObject(t *testing.T) {
	t.Run("create contact", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/objects/contacts", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"id": "67890",
				"properties": {
					"email": "new@example.com",
					"firstname": "New",
					"lastname": "Contact"
				},
				"createdAt": "2024-01-20T10:00:00Z",
				"updatedAt": "2024-01-20T10:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		obj, err := client.CreateObject(ObjectTypeContacts, map[string]interface{}{
			"email":     "new@example.com",
			"firstname": "New",
			"lastname":  "Contact",
		})

		require.NoError(t, err)
		assert.Equal(t, "67890", obj.ID)
		assert.Equal(t, "new@example.com", obj.GetProperty("email"))
	})

	t.Run("bad request - missing required field", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": "error", "message": "Property 'email' is required"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		obj, err := client.CreateObject(ObjectTypeContacts, map[string]interface{}{
			"firstname": "No Email",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
		assert.Nil(t, obj)
	})
}

func TestClient_UpdateObject(t *testing.T) {
	t.Run("update contact", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/objects/contacts/12345", r.URL.Path)
			assert.Equal(t, http.MethodPatch, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "12345",
				"properties": {
					"email": "john@example.com",
					"firstname": "Johnny",
					"lastname": "Doe"
				},
				"createdAt": "2024-01-15T10:00:00Z",
				"updatedAt": "2024-01-20T10:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		obj, err := client.UpdateObject(ObjectTypeContacts, "12345", map[string]interface{}{
			"firstname": "Johnny",
		})

		require.NoError(t, err)
		assert.Equal(t, "12345", obj.ID)
		assert.Equal(t, "Johnny", obj.GetProperty("firstname"))
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		obj, err := client.UpdateObject(ObjectTypeContacts, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "object ID is required")
		assert.Nil(t, obj)
	})
}

func TestClient_DeleteObject(t *testing.T) {
	t.Run("delete contact", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/objects/contacts/12345", r.URL.Path)
			assert.Equal(t, http.MethodDelete, r.Method)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		err := client.DeleteObject(ObjectTypeContacts, "12345")
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Contact not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		err := client.DeleteObject(ObjectTypeContacts, "99999")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		err := client.DeleteObject(ObjectTypeContacts, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "object ID is required")
	})
}

func TestClient_SearchObjects(t *testing.T) {
	t.Run("search contacts by email", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/objects/contacts/search", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "123",
						"properties": {
							"email": "john@example.com",
							"firstname": "John"
						}
					}
				],
				"total": 1
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		result, err := client.SearchObjects(ObjectTypeContacts, SearchRequest{
			FilterGroups: []SearchFilterGroup{
				{
					Filters: []SearchFilter{
						{
							PropertyName: "email",
							Operator:     "EQ",
							Value:        "john@example.com",
						},
					},
				},
			},
			Properties: []string{"email", "firstname"},
			Limit:      10,
		})

		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "123", result.Results[0].ID)
	})

	t.Run("search with sort", func(t *testing.T) {
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

		_, err := client.SearchObjects(ObjectTypeContacts, SearchRequest{
			Sorts: []SearchSort{
				{PropertyName: "createdate", Direction: "DESCENDING"},
			},
		})

		require.NoError(t, err)
	})
}

func TestClient_AllObjectTypes(t *testing.T) {
	// Verify all object types work with the generic methods
	objectTypes := []ObjectType{
		ObjectTypeContacts,
		ObjectTypeCompanies,
		ObjectTypeDeals,
		ObjectTypeTickets,
		ObjectTypeProducts,
		ObjectTypeLineItems,
		ObjectTypeQuotes,
		ObjectTypeNotes,
		ObjectTypeCalls,
		ObjectTypeEmails,
		ObjectTypeMeetings,
		ObjectTypeTasks,
	}

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

	for _, ot := range objectTypes {
		t.Run(string(ot), func(t *testing.T) {
			result, err := client.ListObjects(ot, ListOptions{Limit: 1})
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}
