package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Schema represents a HubSpot custom object schema
type Schema struct {
	ID                         string           `json:"id,omitempty"`
	ObjectTypeID               string           `json:"objectTypeId,omitempty"`
	Name                       string           `json:"name"`
	FullyQualifiedName         string           `json:"fullyQualifiedName,omitempty"`
	Labels                     SchemaLabels     `json:"labels,omitempty"`
	PrimaryDisplayProperty     string           `json:"primaryDisplayProperty,omitempty"`
	SecondaryDisplayProperties []string         `json:"secondaryDisplayProperties,omitempty"`
	RequiredProperties         []string         `json:"requiredProperties,omitempty"`
	SearchableProperties       []string         `json:"searchableProperties,omitempty"`
	Properties                 []SchemaProperty `json:"properties,omitempty"`
	AssociatedObjects          []string         `json:"associatedObjects,omitempty"`
	Archived                   bool             `json:"archived,omitempty"`
	CreatedAt                  string           `json:"createdAt,omitempty"`
	UpdatedAt                  string           `json:"updatedAt,omitempty"`
}

// SchemaLabels represents the labels for a custom object schema
type SchemaLabels struct {
	Singular string `json:"singular,omitempty"`
	Plural   string `json:"plural,omitempty"`
}

// SchemaProperty represents a property definition in a schema
type SchemaProperty struct {
	Name           string `json:"name"`
	Label          string `json:"label,omitempty"`
	Type           string `json:"type"`
	FieldType      string `json:"fieldType,omitempty"`
	Description    string `json:"description,omitempty"`
	GroupName      string `json:"groupName,omitempty"`
	HasUniqueValue bool   `json:"hasUniqueValue,omitempty"`
}

// SchemaList represents a paginated list of schemas
type SchemaList struct {
	Results []Schema `json:"results"`
	Paging  *Paging  `json:"paging,omitempty"`
}

// ListSchemas retrieves custom object schemas with pagination
func (c *Client) ListSchemas(opts ListOptions) (*SchemaList, error) {
	url := fmt.Sprintf("%s/crm/v3/schemas", c.BaseURL)

	params := make(map[string]string)
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
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

	var result SchemaList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse schemas response: %w", err)
	}

	return &result, nil
}

// GetSchema retrieves a single custom object schema by fully qualified name
func (c *Client) GetSchema(fullyQualifiedName string) (*Schema, error) {
	if fullyQualifiedName == "" {
		return nil, fmt.Errorf("fully qualified name is required")
	}

	url := fmt.Sprintf("%s/crm/v3/schemas/%s", c.BaseURL, fullyQualifiedName)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Schema
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse schema response: %w", err)
	}

	return &result, nil
}

// CreateSchema creates a new custom object schema
func (c *Client) CreateSchema(schema map[string]interface{}) (*Schema, error) {
	url := fmt.Sprintf("%s/crm/v3/schemas", c.BaseURL)

	body, err := c.post(url, schema)
	if err != nil {
		return nil, err
	}

	var result Schema
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse schema response: %w", err)
	}

	return &result, nil
}

// DeleteSchema deletes a custom object schema
func (c *Client) DeleteSchema(fullyQualifiedName string) error {
	if fullyQualifiedName == "" {
		return fmt.Errorf("fully qualified name is required")
	}

	url := fmt.Sprintf("%s/crm/v3/schemas/%s", c.BaseURL, fullyQualifiedName)

	_, err := c.delete(url)
	return err
}
