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

func TestClient_IntrospectSchema(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/collector/graphql", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"__schema": {
						"queryType": {"name": "Query"},
						"mutationType": null,
						"subscriptionType": null,
						"types": [
							{
								"kind": "OBJECT",
								"name": "CRM",
								"description": "CRM root type",
								"fields": [
									{
										"name": "contact_collection",
										"description": "List contacts",
										"type": {"kind": "OBJECT", "name": "ContactCollection"},
										"args": [
											{
												"name": "limit",
												"type": {"kind": "SCALAR", "name": "Int"}
											}
										]
									}
								]
							},
							{
								"kind": "OBJECT",
								"name": "Contact",
								"description": "A CRM contact",
								"fields": [
									{
										"name": "id",
										"type": {"kind": "SCALAR", "name": "ID"}
									},
									{
										"name": "email",
										"type": {"kind": "SCALAR", "name": "String"}
									}
								]
							}
						]
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

		schema, err := client.IntrospectSchema()
		require.NoError(t, err)
		assert.NotNil(t, schema.QueryType)
		assert.Equal(t, "Query", *schema.QueryType.Name)
		assert.Len(t, schema.Types, 2)
	})

	t.Run("graphql error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"errors": [{"message": "Introspection not allowed"}]
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		schema, err := client.IntrospectSchema()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Introspection not allowed")
		assert.Nil(t, schema)
	})
}

func TestIntrospectionSchema_GetType(t *testing.T) {
	schema := &IntrospectionSchema{
		Types: []IntrospectionType{
			{Name: "CRM", Kind: "OBJECT"},
			{Name: "Contact", Kind: "OBJECT"},
		},
	}

	t.Run("found", func(t *testing.T) {
		result := schema.GetType("Contact")
		assert.NotNil(t, result)
		assert.Equal(t, "Contact", result.Name)
	})

	t.Run("not found", func(t *testing.T) {
		result := schema.GetType("NonExistent")
		assert.Nil(t, result)
	})
}

func TestIntrospectionSchema_GetRootTypes(t *testing.T) {
	schema := &IntrospectionSchema{
		Types: []IntrospectionType{
			{Name: "CRM", Kind: "OBJECT", Fields: []IntrospectionField{{Name: "contacts"}}},
			{Name: "__Schema", Kind: "OBJECT", Fields: []IntrospectionField{{Name: "types"}}},
			{Name: "String", Kind: "SCALAR"},
			{Name: "Contact", Kind: "OBJECT", Fields: []IntrospectionField{{Name: "id"}}},
		},
	}

	types := schema.GetRootTypes()
	assert.Len(t, types, 2) // CRM and Contact, not __Schema (internal) or String (scalar)

	names := make([]string, len(types))
	for i, typ := range types {
		names[i] = typ.Name
	}
	assert.Contains(t, names, "CRM")
	assert.Contains(t, names, "Contact")
}

func TestIntrospectionTypeRef_TypeName(t *testing.T) {
	t.Run("simple type", func(t *testing.T) {
		name := "String"
		ref := IntrospectionTypeRef{Kind: "SCALAR", Name: &name}
		assert.Equal(t, "String", ref.TypeName())
	})

	t.Run("non-null type", func(t *testing.T) {
		name := "String"
		ref := IntrospectionTypeRef{
			Kind:   "NON_NULL",
			OfType: &IntrospectionTypeRef{Kind: "SCALAR", Name: &name},
		}
		assert.Equal(t, "String!", ref.TypeName())
	})

	t.Run("list type", func(t *testing.T) {
		name := "Contact"
		ref := IntrospectionTypeRef{
			Kind:   "LIST",
			OfType: &IntrospectionTypeRef{Kind: "OBJECT", Name: &name},
		}
		assert.Equal(t, "[Contact]", ref.TypeName())
	})

	t.Run("non-null list", func(t *testing.T) {
		name := "Contact"
		ref := IntrospectionTypeRef{
			Kind: "NON_NULL",
			OfType: &IntrospectionTypeRef{
				Kind:   "LIST",
				OfType: &IntrospectionTypeRef{Kind: "OBJECT", Name: &name},
			},
		}
		assert.Equal(t, "[Contact]!", ref.TypeName())
	})
}
