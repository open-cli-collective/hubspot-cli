package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// File represents a HubSpot file
type File struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Type        string `json:"type"`
	Extension   string `json:"extension"`
	URL         string `json:"url"`
	AccessLevel string `json:"access"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Archived    bool   `json:"archived"`
}

// FileList represents a paginated list of files
type FileList struct {
	Results []File  `json:"results"`
	Paging  *Paging `json:"paging,omitempty"`
}

// Folder represents a HubSpot folder
type Folder struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	ParentID  string `json:"parentFolderId,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Archived  bool   `json:"archived"`
}

// FolderList represents a paginated list of folders
type FolderList struct {
	Results []Folder `json:"results"`
	Paging  *Paging  `json:"paging,omitempty"`
}

// Domain represents a HubSpot domain
type Domain struct {
	ID                   string `json:"id"`
	Domain               string `json:"domain"`
	PrimaryLandingPage   bool   `json:"primaryLandingPage"`
	PrimaryEmail         bool   `json:"primaryEmail"`
	PrimaryBlogPost      bool   `json:"primaryBlogPost"`
	PrimarySitePage      bool   `json:"primarySitePage"`
	PrimaryKnowledge     bool   `json:"primaryKnowledge"`
	IsResolving          bool   `json:"isResolving"`
	IsSslEnabled         bool   `json:"isSslEnabled"`
	IsSslOnly            bool   `json:"isSslOnly"`
	IsUsedForBlogPost    bool   `json:"isUsedForBlogPost"`
	IsUsedForSitePage    bool   `json:"isUsedForSitePage"`
	IsUsedForLandingPage bool   `json:"isUsedForLandingPage"`
	IsUsedForEmail       bool   `json:"isUsedForEmail"`
	IsUsedForKnowledge   bool   `json:"isUsedForKnowledge"`
	CreatedAt            string `json:"createdAt"`
	UpdatedAt            string `json:"updatedAt"`
}

// DomainList represents a paginated list of domains
type DomainList struct {
	Results []Domain `json:"results"`
	Paging  *Paging  `json:"paging,omitempty"`
}

// ListFiles retrieves files with pagination
func (c *Client) ListFiles(opts ListOptions) (*FileList, error) {
	url := fmt.Sprintf("%s/files/v3/files", c.BaseURL)

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

	var result FileList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse files response: %w", err)
	}

	return &result, nil
}

// GetFile retrieves a single file by ID
func (c *Client) GetFile(fileID string) (*File, error) {
	if fileID == "" {
		return nil, fmt.Errorf("file ID is required")
	}

	url := fmt.Sprintf("%s/files/v3/files/%s", c.BaseURL, fileID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result File
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse file response: %w", err)
	}

	return &result, nil
}

// DeleteFile deletes a file by ID
func (c *Client) DeleteFile(fileID string) error {
	if fileID == "" {
		return fmt.Errorf("file ID is required")
	}

	url := fmt.Sprintf("%s/files/v3/files/%s", c.BaseURL, fileID)

	_, err := c.delete(url)
	return err
}

// ListFolders retrieves folders with pagination
func (c *Client) ListFolders(opts ListOptions) (*FolderList, error) {
	url := fmt.Sprintf("%s/files/v3/folders", c.BaseURL)

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

	var result FolderList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse folders response: %w", err)
	}

	return &result, nil
}

// ListDomains retrieves domains with pagination
func (c *Client) ListDomains(opts ListOptions) (*DomainList, error) {
	url := fmt.Sprintf("%s/cms/v3/domains", c.BaseURL)

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

	var result DomainList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse domains response: %w", err)
	}

	return &result, nil
}

// GetDomain retrieves a single domain by ID
func (c *Client) GetDomain(domainID string) (*Domain, error) {
	if domainID == "" {
		return nil, fmt.Errorf("domain ID is required")
	}

	url := fmt.Sprintf("%s/cms/v3/domains/%s", c.BaseURL, domainID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Domain
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse domain response: %w", err)
	}

	return &result, nil
}
