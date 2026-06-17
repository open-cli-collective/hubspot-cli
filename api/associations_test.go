package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListAssociations(t *testing.T) {
	t.Run("results with numeric toObjectId parse successfully", func(t *testing.T) {
		// The HubSpot CRM v4 associations API returns toObjectId as a JSON
		// number, not a string. This used to fail with:
		//   json: cannot unmarshal number into Go struct field
		//   Association.results.toObjectId of type string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/crm/v4/objects/contacts/12345/associations/notes", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"toObjectId": 98765,
						"associationTypes": [
							{
								"category": "HUBSPOT_DEFINED",
								"typeId": 202,
								"label": "Contact to Note"
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

		result, err := client.ListAssociations(ObjectTypeContacts, "12345", ObjectTypeNotes, ListOptions{})
		require.NoError(t, err)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "98765", result.Results[0].ToObjectID.String())
		require.Len(t, result.Results[0].AssociationTypes, 1)
		assert.Equal(t, "HUBSPOT_DEFINED", result.Results[0].AssociationTypes[0].Category)
		assert.Equal(t, 202, result.Results[0].AssociationTypes[0].TypeID)
	})

	t.Run("empty results parse cleanly", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"results": []}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		result, err := client.ListAssociations(ObjectTypeContacts, "12345", ObjectTypeNotes, ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Results)
	})

	t.Run("result re-serializes to valid JSON with numeric id", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": [
					{"toObjectId": 98765, "associationTypes": []}
				]
			}`))
		}))
		defer server.Close()

		client := &Client{
			BaseURL:     server.URL,
			AccessToken: "test-token",
			HTTPClient:  server.Client(),
		}

		result, err := client.ListAssociations(ObjectTypeContacts, "12345", ObjectTypeNotes, ListOptions{})
		require.NoError(t, err)

		// Re-marshalling produces valid JSON; json.Number preserves the
		// number as a number (not a quoted string).
		out, err := json.Marshal(result)
		require.NoError(t, err)

		var roundtrip map[string]interface{}
		require.NoError(t, json.Unmarshal(out, &roundtrip), "marshalled output must be valid JSON")
		assert.Contains(t, string(out), `"toObjectId":98765`)
	})

	t.Run("empty from ID returns error", func(t *testing.T) {
		client := &Client{BaseURL: "https://api.hubapi.com"}
		result, err := client.ListAssociations(ObjectTypeContacts, "", ObjectTypeNotes, ListOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "from object ID is required")
		assert.Nil(t, result)
	})
}
