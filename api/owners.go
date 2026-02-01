package api

import (
	"encoding/json"
	"fmt"
)

// OwnersResponse represents the response from the owners list endpoint
type OwnersResponse struct {
	Results []Owner `json:"results"`
	Paging  *Paging `json:"paging,omitempty"`
}

// GetOwners retrieves all owners (users) from HubSpot
func (c *Client) GetOwners() ([]Owner, error) {
	url := fmt.Sprintf("%s/crm/v3/owners", c.BaseURL)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var resp OwnersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse owners response: %w", err)
	}

	return resp.Results, nil
}

// GetOwner retrieves a single owner by ID
func (c *Client) GetOwner(ownerID string) (*Owner, error) {
	if ownerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	url := fmt.Sprintf("%s/crm/v3/owners/%s", c.BaseURL, ownerID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var owner Owner
	if err := json.Unmarshal(body, &owner); err != nil {
		return nil, fmt.Errorf("failed to parse owner response: %w", err)
	}

	return &owner, nil
}
