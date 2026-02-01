package api

import (
	"encoding/json"
	"fmt"
)

// Association represents a HubSpot association between objects
type Association struct {
	ToObjectID       string            `json:"toObjectId"`
	AssociationTypes []AssociationType `json:"associationTypes"`
}

// AssociationType describes the type of association
type AssociationType struct {
	Category string `json:"category"`
	TypeID   int    `json:"typeId"`
	Label    string `json:"label,omitempty"`
}

// AssociationList represents a paginated list of associations
type AssociationList struct {
	Results []Association `json:"results"`
	Paging  *Paging       `json:"paging,omitempty"`
}

// ListAssociations retrieves associations from one object to another type
// Uses CRM v4 associations API
func (c *Client) ListAssociations(fromType ObjectType, fromID string, toType ObjectType, opts ListOptions) (*AssociationList, error) {
	if fromID == "" {
		return nil, fmt.Errorf("from object ID is required")
	}

	url := fmt.Sprintf("%s/crm/v4/objects/%s/%s/associations/%s", c.BaseURL, fromType, fromID, toType)

	params := make(map[string]string)
	if opts.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", opts.Limit)
	}
	if opts.After != "" {
		params["after"] = opts.After
	}

	if len(params) > 0 {
		url = buildURL(url, params)
	}

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result AssociationList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// CreateAssociation creates an association between two objects
// Uses CRM v4 associations API
func (c *Client) CreateAssociation(fromType ObjectType, fromID string, toType ObjectType, toID string, associationTypeID int) error {
	if fromID == "" {
		return fmt.Errorf("from object ID is required")
	}
	if toID == "" {
		return fmt.Errorf("to object ID is required")
	}

	url := fmt.Sprintf("%s/crm/v4/objects/%s/%s/associations/%s/%s", c.BaseURL, fromType, fromID, toType, toID)

	payload := []map[string]interface{}{
		{
			"associationCategory": "HUBSPOT_DEFINED",
			"associationTypeId":   associationTypeID,
		},
	}

	_, err := c.put(url, payload)
	return err
}

// DeleteAssociation removes an association between two objects
// Uses CRM v4 associations API
func (c *Client) DeleteAssociation(fromType ObjectType, fromID string, toType ObjectType, toID string) error {
	if fromID == "" {
		return fmt.Errorf("from object ID is required")
	}
	if toID == "" {
		return fmt.Errorf("to object ID is required")
	}

	url := fmt.Sprintf("%s/crm/v4/objects/%s/%s/associations/%s/%s", c.BaseURL, fromType, fromID, toType, toID)

	_, err := c.delete(url)
	return err
}
