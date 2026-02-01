package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// HubDBTable represents a HubDB table
type HubDBTable struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Label                 string                 `json:"label,omitempty"`
	Published             bool                   `json:"published,omitempty"`
	RowCount              int                    `json:"rowCount,omitempty"`
	CreatedAt             string                 `json:"createdAt,omitempty"`
	UpdatedAt             string                 `json:"updatedAt,omitempty"`
	PublishedAt           string                 `json:"publishedAt,omitempty"`
	Columns               []HubDBColumn          `json:"columns,omitempty"`
	AllowPublicAPIAccess  bool                   `json:"allowPublicApiAccess,omitempty"`
	AllowChildTables      bool                   `json:"allowChildTables,omitempty"`
	EnableChildTablePages bool                   `json:"enableChildTablePages,omitempty"`
	DynamicMetaTags       map[string]interface{} `json:"dynamicMetaTags,omitempty"`
}

// HubDBColumn represents a column in a HubDB table
type HubDBColumn struct {
	ID          string        `json:"id,omitempty"`
	Name        string        `json:"name"`
	Label       string        `json:"label,omitempty"`
	Type        string        `json:"type"`
	Description string        `json:"description,omitempty"`
	Archived    bool          `json:"archived,omitempty"`
	Options     []HubDBOption `json:"options,omitempty"`
}

// HubDBOption represents an option for select/multiselect columns
type HubDBOption struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Label string `json:"label,omitempty"`
	Order int    `json:"order,omitempty"`
}

// HubDBTableList represents a paginated list of HubDB tables
type HubDBTableList struct {
	Results []HubDBTable `json:"results"`
	Paging  *Paging      `json:"paging,omitempty"`
	Total   int          `json:"total,omitempty"`
}

// HubDBRow represents a row in a HubDB table
type HubDBRow struct {
	ID        string                 `json:"id"`
	Path      string                 `json:"path,omitempty"`
	Name      string                 `json:"name,omitempty"`
	CreatedAt string                 `json:"createdAt,omitempty"`
	UpdatedAt string                 `json:"updatedAt,omitempty"`
	Values    map[string]interface{} `json:"values,omitempty"`
}

// HubDBRowList represents a paginated list of HubDB rows
type HubDBRowList struct {
	Results []HubDBRow `json:"results"`
	Paging  *Paging    `json:"paging,omitempty"`
	Total   int        `json:"total,omitempty"`
}

// ListHubDBTables retrieves HubDB tables with pagination
func (c *Client) ListHubDBTables(opts ListOptions) (*HubDBTableList, error) {
	url := fmt.Sprintf("%s/cms/v3/hubdb/tables", c.BaseURL)

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

	var result HubDBTableList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse hubdb tables response: %w", err)
	}

	return &result, nil
}

// GetHubDBTable retrieves a single HubDB table by ID or name
func (c *Client) GetHubDBTable(tableIDOrName string) (*HubDBTable, error) {
	if tableIDOrName == "" {
		return nil, fmt.Errorf("table ID or name is required")
	}

	url := fmt.Sprintf("%s/cms/v3/hubdb/tables/%s", c.BaseURL, tableIDOrName)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result HubDBTable
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse hubdb table response: %w", err)
	}

	return &result, nil
}

// CreateHubDBTable creates a new HubDB table
func (c *Client) CreateHubDBTable(table map[string]interface{}) (*HubDBTable, error) {
	url := fmt.Sprintf("%s/cms/v3/hubdb/tables", c.BaseURL)

	body, err := c.post(url, table)
	if err != nil {
		return nil, err
	}

	var result HubDBTable
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse hubdb table response: %w", err)
	}

	return &result, nil
}

// DeleteHubDBTable deletes a HubDB table
func (c *Client) DeleteHubDBTable(tableIDOrName string) error {
	if tableIDOrName == "" {
		return fmt.Errorf("table ID or name is required")
	}

	url := fmt.Sprintf("%s/cms/v3/hubdb/tables/%s", c.BaseURL, tableIDOrName)

	_, err := c.delete(url)
	return err
}

// PublishHubDBTable publishes a HubDB table draft
func (c *Client) PublishHubDBTable(tableIDOrName string) (*HubDBTable, error) {
	if tableIDOrName == "" {
		return nil, fmt.Errorf("table ID or name is required")
	}

	url := fmt.Sprintf("%s/cms/v3/hubdb/tables/%s/draft/publish", c.BaseURL, tableIDOrName)

	body, err := c.post(url, nil)
	if err != nil {
		return nil, err
	}

	var result HubDBTable
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse hubdb table response: %w", err)
	}

	return &result, nil
}

// ListHubDBRows retrieves rows from a HubDB table
func (c *Client) ListHubDBRows(tableIDOrName string, opts ListOptions) (*HubDBRowList, error) {
	if tableIDOrName == "" {
		return nil, fmt.Errorf("table ID or name is required")
	}

	url := fmt.Sprintf("%s/cms/v3/hubdb/tables/%s/rows", c.BaseURL, tableIDOrName)

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

	var result HubDBRowList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse hubdb rows response: %w", err)
	}

	return &result, nil
}

// GetHubDBRow retrieves a single row from a HubDB table
func (c *Client) GetHubDBRow(tableIDOrName, rowID string) (*HubDBRow, error) {
	if tableIDOrName == "" {
		return nil, fmt.Errorf("table ID or name is required")
	}
	if rowID == "" {
		return nil, fmt.Errorf("row ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/hubdb/tables/%s/rows/%s", c.BaseURL, tableIDOrName, rowID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result HubDBRow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse hubdb row response: %w", err)
	}

	return &result, nil
}

// CreateHubDBRow creates a new row in a HubDB table draft
func (c *Client) CreateHubDBRow(tableIDOrName string, row map[string]interface{}) (*HubDBRow, error) {
	if tableIDOrName == "" {
		return nil, fmt.Errorf("table ID or name is required")
	}

	url := fmt.Sprintf("%s/cms/v3/hubdb/tables/%s/rows/draft", c.BaseURL, tableIDOrName)

	body, err := c.post(url, row)
	if err != nil {
		return nil, err
	}

	var result HubDBRow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse hubdb row response: %w", err)
	}

	return &result, nil
}

// UpdateHubDBRow updates a row in a HubDB table draft
func (c *Client) UpdateHubDBRow(tableIDOrName, rowID string, updates map[string]interface{}) (*HubDBRow, error) {
	if tableIDOrName == "" {
		return nil, fmt.Errorf("table ID or name is required")
	}
	if rowID == "" {
		return nil, fmt.Errorf("row ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/hubdb/tables/%s/rows/draft/%s", c.BaseURL, tableIDOrName, rowID)

	body, err := c.patch(url, updates)
	if err != nil {
		return nil, err
	}

	var result HubDBRow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse hubdb row response: %w", err)
	}

	return &result, nil
}

// DeleteHubDBRow deletes a row from a HubDB table draft
func (c *Client) DeleteHubDBRow(tableIDOrName, rowID string) error {
	if tableIDOrName == "" {
		return fmt.Errorf("table ID or name is required")
	}
	if rowID == "" {
		return fmt.Errorf("row ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/hubdb/tables/%s/rows/draft/%s", c.BaseURL, tableIDOrName, rowID)

	_, err := c.delete(url)
	return err
}
