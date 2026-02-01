package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// PageType represents the type of CMS page
type PageType string

const (
	PageTypeSite    PageType = "site-pages"
	PageTypeLanding PageType = "landing-pages"
)

// Page represents a HubSpot CMS page
type Page struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Slug            string                 `json:"slug,omitempty"`
	State           string                 `json:"state,omitempty"`
	AuthorName      string                 `json:"authorName,omitempty"`
	Domain          string                 `json:"domain,omitempty"`
	Subcategory     string                 `json:"subcategory,omitempty"`
	HTMLTitle       string                 `json:"htmlTitle,omitempty"`
	MetaDescription string                 `json:"metaDescription,omitempty"`
	PublishDate     string                 `json:"publishDate,omitempty"`
	CreatedAt       string                 `json:"created,omitempty"`
	UpdatedAt       string                 `json:"updated,omitempty"`
	Archived        bool                   `json:"archived,omitempty"`
	ArchivedAt      string                 `json:"archivedAt,omitempty"`
	CurrentState    string                 `json:"currentState,omitempty"`
	LayoutSections  map[string]interface{} `json:"layoutSections,omitempty"`
}

// PageList represents a paginated list of pages
type PageList struct {
	Results []Page  `json:"results"`
	Paging  *Paging `json:"paging,omitempty"`
	Total   int     `json:"total,omitempty"`
}

// ListPages retrieves pages with pagination
func (c *Client) ListPages(pageType PageType, opts ListOptions) (*PageList, error) {
	url := fmt.Sprintf("%s/cms/v3/pages/%s", c.BaseURL, pageType)

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

	var result PageList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pages response: %w", err)
	}

	return &result, nil
}

// GetPage retrieves a single page by ID
func (c *Client) GetPage(pageType PageType, pageID string) (*Page, error) {
	if pageID == "" {
		return nil, fmt.Errorf("page ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/pages/%s/%s", c.BaseURL, pageType, pageID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Page
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}

	return &result, nil
}

// CreatePage creates a new page
func (c *Client) CreatePage(pageType PageType, page map[string]interface{}) (*Page, error) {
	url := fmt.Sprintf("%s/cms/v3/pages/%s", c.BaseURL, pageType)

	body, err := c.post(url, page)
	if err != nil {
		return nil, err
	}

	var result Page
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}

	return &result, nil
}

// UpdatePage updates an existing page
func (c *Client) UpdatePage(pageType PageType, pageID string, updates map[string]interface{}) (*Page, error) {
	if pageID == "" {
		return nil, fmt.Errorf("page ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/pages/%s/%s", c.BaseURL, pageType, pageID)

	body, err := c.patch(url, updates)
	if err != nil {
		return nil, err
	}

	var result Page
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}

	return &result, nil
}

// DeletePage archives a page
func (c *Client) DeletePage(pageType PageType, pageID string) error {
	if pageID == "" {
		return fmt.Errorf("page ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/pages/%s/%s", c.BaseURL, pageType, pageID)

	_, err := c.delete(url)
	return err
}

// ClonePage creates a copy of an existing page
func (c *Client) ClonePage(pageType PageType, pageID string) (*Page, error) {
	if pageID == "" {
		return nil, fmt.Errorf("page ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/pages/%s/%s/clone", c.BaseURL, pageType, pageID)

	body, err := c.post(url, nil)
	if err != nil {
		return nil, err
	}

	var result Page
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}

	return &result, nil
}
