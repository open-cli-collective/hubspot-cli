package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListForms(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/marketing/v3/forms", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "form-123",
						"name": "Contact Us",
						"formType": "hubspot",
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

		result, err := client.ListForms(ListOptions{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "form-123", result.Results[0].ID)
		assert.Equal(t, "Contact Us", result.Results[0].Name)
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

		result, err := client.ListForms(ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})
}

func TestClient_GetForm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/marketing/v3/forms/form-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "form-123",
				"name": "Contact Us",
				"formType": "hubspot",
				"createdAt": "2024-01-15T10:00:00Z",
				"updatedAt": "2024-01-16T12:00:00Z",
				"archived": false,
				"fieldGroups": [
					{
						"groupType": "default_group",
						"fields": [
							{
								"name": "email",
								"label": "Email",
								"fieldType": "email",
								"required": true,
								"hidden": false
							}
						]
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

		form, err := client.GetForm("form-123")
		require.NoError(t, err)
		assert.Equal(t, "form-123", form.ID)
		assert.Equal(t, "Contact Us", form.Name)
		assert.Len(t, form.FieldGroups, 1)
		assert.Len(t, form.FieldGroups[0].Fields, 1)
		assert.Equal(t, "email", form.FieldGroups[0].Fields[0].Name)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Form not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		form, err := client.GetForm("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, form)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		form, err := client.GetForm("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "form ID is required")
		assert.Nil(t, form)
	})
}

func TestClient_GetFormSubmissions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/marketing/v3/forms/form-123/submissions", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "sub-456",
						"submittedAt": "2024-01-20T10:00:00Z",
						"values": {
							"email": "john@example.com",
							"firstname": "John"
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

		result, err := client.GetFormSubmissions("form-123", ListOptions{})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "sub-456", result.Results[0].ID)
		assert.Equal(t, "john@example.com", result.Results[0].Values["email"])
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		result, err := client.GetFormSubmissions("", ListOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "form ID is required")
		assert.Nil(t, result)
	})
}

func TestClient_ListCampaigns(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/marketing/v3/campaigns", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": "campaign-123",
						"name": "Q1 Launch",
						"createdAt": "2024-01-01T10:00:00Z",
						"updatedAt": "2024-01-15T12:00:00Z"
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

		result, err := client.ListCampaigns(ListOptions{})
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, "campaign-123", result.Results[0].ID)
		assert.Equal(t, "Q1 Launch", result.Results[0].Name)
	})
}

func TestClient_GetCampaign(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/marketing/v3/campaigns/campaign-123", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "campaign-123",
				"name": "Q1 Launch",
				"createdAt": "2024-01-01T10:00:00Z",
				"updatedAt": "2024-01-15T12:00:00Z"
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		campaign, err := client.GetCampaign("campaign-123")
		require.NoError(t, err)
		assert.Equal(t, "campaign-123", campaign.ID)
		assert.Equal(t, "Q1 Launch", campaign.Name)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Campaign not found"}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		campaign, err := client.GetCampaign("nonexistent")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, campaign)
	})

	t.Run("empty ID", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		campaign, err := client.GetCampaign("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "campaign ID is required")
		assert.Nil(t, campaign)
	})
}
