package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetOwners(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/owners", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "12345",
						"email": "john@example.com",
						"firstName": "John",
						"lastName": "Doe",
						"userId": 67890,
						"archived": false
					},
					{
						"id": "12346",
						"email": "jane@example.com",
						"firstName": "Jane",
						"lastName": "Smith",
						"userId": 67891,
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

		owners, err := client.GetOwners()
		require.NoError(t, err)
		assert.Len(t, owners, 2)

		assert.Equal(t, "12345", owners[0].ID)
		assert.Equal(t, "john@example.com", owners[0].Email)
		assert.Equal(t, "John", owners[0].FirstName)
		assert.Equal(t, "Doe", owners[0].LastName)

		assert.Equal(t, "12346", owners[1].ID)
		assert.Equal(t, "jane@example.com", owners[1].Email)
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

		owners, err := client.GetOwners()
		require.NoError(t, err)
		assert.Empty(t, owners)
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status": "error", "message": "Invalid access token"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "bad-token",
			HTTPClient:  server.Client(),
		}

		owners, err := client.GetOwners()
		assert.Error(t, err)
		assert.True(t, IsUnauthorized(err))
		assert.Nil(t, owners)
	})
}

func TestClient_GetOwner(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v3/owners/12345", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "12345",
				"email": "john@example.com",
				"firstName": "John",
				"lastName": "Doe",
				"userId": 67890,
				"archived": false,
				"teams": [
					{
						"id": "team1",
						"name": "Sales",
						"primary": true
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

		owner, err := client.GetOwner("12345")
		require.NoError(t, err)
		assert.Equal(t, "12345", owner.ID)
		assert.Equal(t, "john@example.com", owner.Email)
		assert.Equal(t, "John", owner.FirstName)
		assert.Equal(t, "Doe", owner.LastName)
		assert.Len(t, owner.Teams, 1)
		assert.Equal(t, "Sales", owner.Teams[0].Name)
		assert.True(t, owner.Teams[0].Primary)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Owner not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		owner, err := client.GetOwner("99999")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, owner)
	})

	t.Run("empty owner ID", func(t *testing.T) {
		client := &Client{
			BaseURL:     "https://api.hubapi.com",
			AccessToken: "test-token",
		}

		owner, err := client.GetOwner("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "owner ID is required")
		assert.Nil(t, owner)
	})
}

func TestOwner_FullName(t *testing.T) {
	tests := []struct {
		name  string
		owner Owner
		want  string
	}{
		{
			name: "both names",
			owner: Owner{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			want: "John Doe",
		},
		{
			name: "first name only",
			owner: Owner{
				FirstName: "John",
				Email:     "john@example.com",
			},
			want: "John",
		},
		{
			name: "last name only",
			owner: Owner{
				LastName: "Doe",
				Email:    "john@example.com",
			},
			want: "Doe",
		},
		{
			name: "no names - falls back to email",
			owner: Owner{
				Email: "john@example.com",
			},
			want: "john@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.owner.FullName())
		})
	}
}
