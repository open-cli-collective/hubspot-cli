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

// Channel represents a HubSpot conversations channel
type Channel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	AccountID string `json:"accountId,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ChannelList represents a paginated list of channels
type ChannelList struct {
	Results []Channel `json:"results"`
	Paging  *Paging   `json:"paging,omitempty"`
}

// Message represents a HubSpot conversations message
type Message struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Text        string                 `json:"text,omitempty"`
	RichText    string                 `json:"richText,omitempty"`
	Direction   string                 `json:"direction,omitempty"`
	Status      string                 `json:"status,omitempty"`
	ChannelID   string                 `json:"channelId,omitempty"`
	ChannelType string                 `json:"channelAccountId,omitempty"`
	SenderID    string                 `json:"senderId,omitempty"`
	Recipients  []MessageRecipient     `json:"recipients,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
	Client      map[string]interface{} `json:"client,omitempty"`
}

// MessageRecipient represents a recipient of a message
type MessageRecipient struct {
	RecipientID string `json:"recipientId,omitempty"`
}

// MessageList represents a paginated list of messages
type MessageList struct {
	Results []Message `json:"results"`
	Paging  *Paging   `json:"paging,omitempty"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	RichText  string `json:"richText,omitempty"`
	SenderID  string `json:"senderActorId,omitempty"`
	ChannelID string `json:"channelId,omitempty"`
}

// ListChannels retrieves channels with pagination
func (c *Client) ListChannels(opts ListOptions) (*ChannelList, error) {
	url := fmt.Sprintf("%s/conversations/v3/conversations/channels", c.BaseURL)

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

	var result ChannelList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse channels response: %w", err)
	}

	return &result, nil
}

// GetChannel retrieves a single channel by ID
func (c *Client) GetChannel(channelID string) (*Channel, error) {
	if channelID == "" {
		return nil, fmt.Errorf("channel ID is required")
	}

	url := fmt.Sprintf("%s/conversations/v3/conversations/channels/%s", c.BaseURL, channelID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Channel
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse channel response: %w", err)
	}

	return &result, nil
}

// ListMessages retrieves messages for a thread
func (c *Client) ListMessages(threadID string, opts ListOptions) (*MessageList, error) {
	if threadID == "" {
		return nil, fmt.Errorf("thread ID is required")
	}

	url := fmt.Sprintf("%s/conversations/v3/conversations/threads/%s/messages", c.BaseURL, threadID)

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

	var result MessageList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse messages response: %w", err)
	}

	return &result, nil
}

// SendMessage sends a message to a thread
func (c *Client) SendMessage(threadID string, req SendMessageRequest) (*Message, error) {
	if threadID == "" {
		return nil, fmt.Errorf("thread ID is required")
	}

	url := fmt.Sprintf("%s/conversations/v3/conversations/threads/%s/messages", c.BaseURL, threadID)

	body, err := c.post(url, req)
	if err != nil {
		return nil, err
	}

	var result Message
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse message response: %w", err)
	}

	return &result, nil
}
