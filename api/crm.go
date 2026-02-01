package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ObjectType represents a HubSpot CRM object type
type ObjectType string

// Standard CRM object types
const (
	ObjectTypeContacts  ObjectType = "contacts"
	ObjectTypeCompanies ObjectType = "companies"
	ObjectTypeDeals     ObjectType = "deals"
	ObjectTypeTickets   ObjectType = "tickets"
	ObjectTypeProducts  ObjectType = "products"
	ObjectTypeLineItems ObjectType = "line_items"
	ObjectTypeQuotes    ObjectType = "quotes"
	ObjectTypeNotes     ObjectType = "notes"
	ObjectTypeCalls     ObjectType = "calls"
	ObjectTypeEmails    ObjectType = "emails"
	ObjectTypeMeetings  ObjectType = "meetings"
	ObjectTypeTasks     ObjectType = "tasks"
)

// CRMObject represents a generic HubSpot CRM object
type CRMObject struct {
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
	Archived   bool                   `json:"archived,omitempty"`
}

// GetProperty returns a property value as a string, or empty string if not found
func (o *CRMObject) GetProperty(name string) string {
	if o.Properties == nil {
		return ""
	}
	if v, ok := o.Properties[name]; ok {
		if v == nil {
			return ""
		}
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(val)
		default:
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

// CRMObjectList represents a paginated list of CRM objects
type CRMObjectList struct {
	Results []CRMObject `json:"results"`
	Paging  *Paging     `json:"paging,omitempty"`
}

// ListOptions contains common options for list operations
type ListOptions struct {
	Limit      int
	After      string
	Properties []string
}

// SearchFilter represents a single filter condition
type SearchFilter struct {
	PropertyName string `json:"propertyName"`
	Operator     string `json:"operator"`
	Value        string `json:"value,omitempty"`
}

// SearchFilterGroup represents a group of filters (ANDed together)
type SearchFilterGroup struct {
	Filters []SearchFilter `json:"filters"`
}

// SearchSort represents a sort condition
type SearchSort struct {
	PropertyName string `json:"propertyName"`
	Direction    string `json:"direction"` // "ASCENDING" or "DESCENDING"
}

// SearchRequest represents a CRM search request
type SearchRequest struct {
	FilterGroups []SearchFilterGroup `json:"filterGroups,omitempty"`
	Sorts        []SearchSort        `json:"sorts,omitempty"`
	Properties   []string            `json:"properties,omitempty"`
	Limit        int                 `json:"limit,omitempty"`
	After        string              `json:"after,omitempty"`
}

// CreateRequest represents a CRM object creation request
type CreateRequest struct {
	Properties map[string]interface{} `json:"properties"`
}

// UpdateRequest represents a CRM object update request
type UpdateRequest struct {
	Properties map[string]interface{} `json:"properties"`
}

// ListObjects lists CRM objects of the given type
func (c *Client) ListObjects(objectType ObjectType, opts ListOptions) (*CRMObjectList, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/%s", c.BaseURL, objectType)

	params := make(map[string]string)
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.After != "" {
		params["after"] = opts.After
	}
	if len(opts.Properties) > 0 {
		// HubSpot accepts comma-separated properties
		props := ""
		for i, p := range opts.Properties {
			if i > 0 {
				props += ","
			}
			props += p
		}
		params["properties"] = props
	}

	if len(params) > 0 {
		url = buildURL(url, params)
	}

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result CRMObjectList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetObject retrieves a single CRM object by ID
func (c *Client) GetObject(objectType ObjectType, id string, properties []string) (*CRMObject, error) {
	if id == "" {
		return nil, fmt.Errorf("object ID is required")
	}

	url := fmt.Sprintf("%s/crm/v3/objects/%s/%s", c.BaseURL, objectType, id)

	params := make(map[string]string)
	if len(properties) > 0 {
		props := ""
		for i, p := range properties {
			if i > 0 {
				props += ","
			}
			props += p
		}
		params["properties"] = props
	}

	if len(params) > 0 {
		url = buildURL(url, params)
	}

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result CRMObject
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// CreateObject creates a new CRM object
func (c *Client) CreateObject(objectType ObjectType, properties map[string]interface{}) (*CRMObject, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/%s", c.BaseURL, objectType)

	req := CreateRequest{Properties: properties}

	body, err := c.post(url, req)
	if err != nil {
		return nil, err
	}

	var result CRMObject
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// UpdateObject updates an existing CRM object
func (c *Client) UpdateObject(objectType ObjectType, id string, properties map[string]interface{}) (*CRMObject, error) {
	if id == "" {
		return nil, fmt.Errorf("object ID is required")
	}

	url := fmt.Sprintf("%s/crm/v3/objects/%s/%s", c.BaseURL, objectType, id)

	req := UpdateRequest{Properties: properties}

	body, err := c.patch(url, req)
	if err != nil {
		return nil, err
	}

	var result CRMObject
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// DeleteObject deletes a CRM object (moves to archive)
func (c *Client) DeleteObject(objectType ObjectType, id string) error {
	if id == "" {
		return fmt.Errorf("object ID is required")
	}

	url := fmt.Sprintf("%s/crm/v3/objects/%s/%s", c.BaseURL, objectType, id)

	_, err := c.delete(url)
	return err
}

// SearchObjects searches for CRM objects
func (c *Client) SearchObjects(objectType ObjectType, req SearchRequest) (*CRMObjectList, error) {
	url := fmt.Sprintf("%s/crm/v3/objects/%s/search", c.BaseURL, objectType)

	body, err := c.post(url, req)
	if err != nil {
		return nil, err
	}

	var result CRMObjectList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
