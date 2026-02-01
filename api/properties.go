package api

import (
	"encoding/json"
	"fmt"
)

// Property represents a HubSpot property definition
type Property struct {
	Name                 string            `json:"name"`
	Label                string            `json:"label"`
	Type                 string            `json:"type"`
	FieldType            string            `json:"fieldType"`
	Description          string            `json:"description,omitempty"`
	GroupName            string            `json:"groupName"`
	Options              []PropertyOption  `json:"options,omitempty"`
	DisplayOrder         int               `json:"displayOrder,omitempty"`
	Calculated           bool              `json:"calculated,omitempty"`
	ExternalOptions      bool              `json:"externalOptions,omitempty"`
	HasUniqueValue       bool              `json:"hasUniqueValue,omitempty"`
	Hidden               bool              `json:"hidden,omitempty"`
	HubspotDefined       bool              `json:"hubspotDefined,omitempty"`
	ModificationMetadata *ModificationMeta `json:"modificationMetadata,omitempty"`
	FormField            bool              `json:"formField,omitempty"`
	CreatedAt            string            `json:"createdAt,omitempty"`
	UpdatedAt            string            `json:"updatedAt,omitempty"`
	ArchivedAt           string            `json:"archivedAt,omitempty"`
	Archived             bool              `json:"archived,omitempty"`
}

// PropertyOption represents an option for enumeration properties
type PropertyOption struct {
	Label        string `json:"label"`
	Value        string `json:"value"`
	Description  string `json:"description,omitempty"`
	DisplayOrder int    `json:"displayOrder,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`
}

// ModificationMeta contains metadata about property modifications
type ModificationMeta struct {
	Archivable         bool `json:"archivable"`
	ReadOnlyValue      bool `json:"readOnlyValue"`
	ReadOnlyDefinition bool `json:"readOnlyDefinition"`
}

// PropertyList represents a list of properties
type PropertyList struct {
	Results []Property `json:"results"`
}

// CreatePropertyRequest represents a request to create a property
type CreatePropertyRequest struct {
	Name           string           `json:"name"`
	Label          string           `json:"label"`
	Type           string           `json:"type"`
	FieldType      string           `json:"fieldType"`
	GroupName      string           `json:"groupName"`
	Description    string           `json:"description,omitempty"`
	Options        []PropertyOption `json:"options,omitempty"`
	DisplayOrder   int              `json:"displayOrder,omitempty"`
	HasUniqueValue bool             `json:"hasUniqueValue,omitempty"`
	Hidden         bool             `json:"hidden,omitempty"`
	FormField      bool             `json:"formField,omitempty"`
}

// ListProperties lists all properties for an object type
func (c *Client) ListProperties(objectType ObjectType) (*PropertyList, error) {
	url := fmt.Sprintf("%s/crm/v3/properties/%s", c.BaseURL, objectType)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result PropertyList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetProperty retrieves a specific property by name
func (c *Client) GetProperty(objectType ObjectType, propertyName string) (*Property, error) {
	if propertyName == "" {
		return nil, fmt.Errorf("property name is required")
	}

	url := fmt.Sprintf("%s/crm/v3/properties/%s/%s", c.BaseURL, objectType, propertyName)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Property
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// CreateProperty creates a new property for an object type
func (c *Client) CreateProperty(objectType ObjectType, req CreatePropertyRequest) (*Property, error) {
	url := fmt.Sprintf("%s/crm/v3/properties/%s", c.BaseURL, objectType)

	body, err := c.post(url, req)
	if err != nil {
		return nil, err
	}

	var result Property
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// DeleteProperty archives (soft deletes) a property
func (c *Client) DeleteProperty(objectType ObjectType, propertyName string) error {
	if propertyName == "" {
		return fmt.Errorf("property name is required")
	}

	url := fmt.Sprintf("%s/crm/v3/properties/%s/%s", c.BaseURL, objectType, propertyName)

	_, err := c.delete(url)
	return err
}
