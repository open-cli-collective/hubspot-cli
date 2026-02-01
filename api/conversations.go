package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Inbox represents a HubSpot conversations inbox
type Inbox struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Archived  bool   `json:"archived"`
}

// InboxList represents a paginated list of inboxes
type InboxList struct {
	Results []Inbox `json:"results"`
	Paging  *Paging `json:"paging,omitempty"`
}

// Thread represents a HubSpot conversations thread
type Thread struct {
	ID                  string `json:"id"`
	Status              string `json:"status"`
	AssociatedContactID string `json:"associatedContactId,omitempty"`
	InboxID             string `json:"inboxId,omitempty"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
	ClosedAt            string `json:"closedAt,omitempty"`
	Archived            bool   `json:"archived"`
}

// ThreadList represents a paginated list of threads
type ThreadList struct {
	Results []Thread `json:"results"`
	Paging  *Paging  `json:"paging,omitempty"`
}

// ListInboxes retrieves inboxes with pagination
func (c *Client) ListInboxes(opts ListOptions) (*InboxList, error) {
	url := fmt.Sprintf("%s/conversations/v3/conversations/inboxes", c.BaseURL)

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

	var result InboxList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse inboxes response: %w", err)
	}

	return &result, nil
}

// GetInbox retrieves a single inbox by ID
func (c *Client) GetInbox(inboxID string) (*Inbox, error) {
	if inboxID == "" {
		return nil, fmt.Errorf("inbox ID is required")
	}

	url := fmt.Sprintf("%s/conversations/v3/conversations/inboxes/%s", c.BaseURL, inboxID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Inbox
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse inbox response: %w", err)
	}

	return &result, nil
}

// ListThreads retrieves threads with pagination
func (c *Client) ListThreads(opts ListOptions) (*ThreadList, error) {
	url := fmt.Sprintf("%s/conversations/v3/conversations/threads", c.BaseURL)

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

	var result ThreadList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse threads response: %w", err)
	}

	return &result, nil
}

// GetThread retrieves a single thread by ID
func (c *Client) GetThread(threadID string) (*Thread, error) {
	if threadID == "" {
		return nil, fmt.Errorf("thread ID is required")
	}

	url := fmt.Sprintf("%s/conversations/v3/conversations/threads/%s", c.BaseURL, threadID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Thread
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse thread response: %w", err)
	}

	return &result, nil
}
