package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Form represents a HubSpot form
type Form struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	FormType    string           `json:"formType"`
	CreatedAt   string           `json:"createdAt"`
	UpdatedAt   string           `json:"updatedAt"`
	Archived    bool             `json:"archived"`
	FieldGroups []FormFieldGroup `json:"fieldGroups,omitempty"`
}

// FormFieldGroup represents a group of fields in a form
type FormFieldGroup struct {
	GroupType    string      `json:"groupType"`
	RichTextType string      `json:"richTextType,omitempty"`
	Fields       []FormField `json:"fields,omitempty"`
}

// FormField represents a field in a form
type FormField struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	FieldType string `json:"fieldType"`
	Required  bool   `json:"required"`
	Hidden    bool   `json:"hidden"`
}

// FormList represents a paginated list of forms
type FormList struct {
	Results []Form  `json:"results"`
	Paging  *Paging `json:"paging,omitempty"`
}

// FormSubmission represents a form submission
type FormSubmission struct {
	ID          string                 `json:"id"`
	SubmittedAt string                 `json:"submittedAt"`
	Values      map[string]interface{} `json:"values"`
}

// FormSubmissionList represents a paginated list of form submissions
type FormSubmissionList struct {
	Results []FormSubmission `json:"results"`
	Paging  *Paging          `json:"paging,omitempty"`
}

// Campaign represents a HubSpot marketing campaign
type Campaign struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CampaignList represents a paginated list of campaigns
type CampaignList struct {
	Results []Campaign `json:"results"`
	Paging  *Paging    `json:"paging,omitempty"`
}

// ListForms retrieves all forms with pagination
func (c *Client) ListForms(opts ListOptions) (*FormList, error) {
	url := fmt.Sprintf("%s/marketing/v3/forms", c.BaseURL)

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

	var result FormList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse forms response: %w", err)
	}

	return &result, nil
}

// GetForm retrieves a single form by ID
func (c *Client) GetForm(formID string) (*Form, error) {
	if formID == "" {
		return nil, fmt.Errorf("form ID is required")
	}

	url := fmt.Sprintf("%s/marketing/v3/forms/%s", c.BaseURL, formID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Form
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse form response: %w", err)
	}

	return &result, nil
}

// GetFormSubmissions retrieves submissions for a form
func (c *Client) GetFormSubmissions(formID string, opts ListOptions) (*FormSubmissionList, error) {
	if formID == "" {
		return nil, fmt.Errorf("form ID is required")
	}

	url := fmt.Sprintf("%s/marketing/v3/forms/%s/submissions", c.BaseURL, formID)

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

	var result FormSubmissionList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse submissions response: %w", err)
	}

	return &result, nil
}

// ListCampaigns retrieves all campaigns with pagination
func (c *Client) ListCampaigns(opts ListOptions) (*CampaignList, error) {
	url := fmt.Sprintf("%s/marketing/v3/campaigns", c.BaseURL)

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

	var result CampaignList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse campaigns response: %w", err)
	}

	return &result, nil
}

// GetCampaign retrieves a single campaign by ID
func (c *Client) GetCampaign(campaignID string) (*Campaign, error) {
	if campaignID == "" {
		return nil, fmt.Errorf("campaign ID is required")
	}

	url := fmt.Sprintf("%s/marketing/v3/campaigns/%s", c.BaseURL, campaignID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Campaign
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse campaign response: %w", err)
	}

	return &result, nil
}

// MarketingEmail represents a HubSpot marketing email
type MarketingEmail struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Subject     string                 `json:"subject,omitempty"`
	Type        string                 `json:"type,omitempty"`
	State       string                 `json:"state,omitempty"`
	PublishDate string                 `json:"publishDate,omitempty"`
	CreatedAt   string                 `json:"created,omitempty"`
	UpdatedAt   string                 `json:"updated,omitempty"`
	Archived    bool                   `json:"archived,omitempty"`
	CampaignID  string                 `json:"campaign,omitempty"`
	FromName    string                 `json:"fromName,omitempty"`
	ReplyTo     string                 `json:"replyTo,omitempty"`
	Content     map[string]interface{} `json:"content,omitempty"`
}

// MarketingEmailList represents a paginated list of marketing emails
type MarketingEmailList struct {
	Results []MarketingEmail `json:"results"`
	Paging  *Paging          `json:"paging,omitempty"`
	Total   int              `json:"total,omitempty"`
}

// ListMarketingEmails retrieves marketing emails with pagination
func (c *Client) ListMarketingEmails(opts ListOptions) (*MarketingEmailList, error) {
	url := fmt.Sprintf("%s/marketing/v3/emails", c.BaseURL)

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

	var result MarketingEmailList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse marketing emails response: %w", err)
	}

	return &result, nil
}

// GetMarketingEmail retrieves a single marketing email by ID
func (c *Client) GetMarketingEmail(emailID string) (*MarketingEmail, error) {
	if emailID == "" {
		return nil, fmt.Errorf("email ID is required")
	}

	url := fmt.Sprintf("%s/marketing/v3/emails/%s", c.BaseURL, emailID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result MarketingEmail
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse marketing email response: %w", err)
	}

	return &result, nil
}

// CreateMarketingEmail creates a new marketing email
func (c *Client) CreateMarketingEmail(email map[string]interface{}) (*MarketingEmail, error) {
	url := fmt.Sprintf("%s/marketing/v3/emails", c.BaseURL)

	body, err := c.post(url, email)
	if err != nil {
		return nil, err
	}

	var result MarketingEmail
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse marketing email response: %w", err)
	}

	return &result, nil
}

// UpdateMarketingEmail updates an existing marketing email
func (c *Client) UpdateMarketingEmail(emailID string, updates map[string]interface{}) (*MarketingEmail, error) {
	if emailID == "" {
		return nil, fmt.Errorf("email ID is required")
	}

	url := fmt.Sprintf("%s/marketing/v3/emails/%s", c.BaseURL, emailID)

	body, err := c.patch(url, updates)
	if err != nil {
		return nil, err
	}

	var result MarketingEmail
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse marketing email response: %w", err)
	}

	return &result, nil
}

// DeleteMarketingEmail archives a marketing email
func (c *Client) DeleteMarketingEmail(emailID string) error {
	if emailID == "" {
		return fmt.Errorf("email ID is required")
	}

	url := fmt.Sprintf("%s/marketing/v3/emails/%s", c.BaseURL, emailID)

	_, err := c.delete(url)
	return err
}
