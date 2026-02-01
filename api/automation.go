package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Workflow represents a HubSpot automation workflow
type Workflow struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Enabled      bool   `json:"isEnabled"`
	ObjectTypeID string `json:"objectTypeId,omitempty"`
	RevisionID   string `json:"revisionId,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

// WorkflowList represents a paginated list of workflows
type WorkflowList struct {
	Results []Workflow `json:"results"`
	Paging  *Paging    `json:"paging,omitempty"`
}

// ListWorkflows retrieves workflows with pagination
func (c *Client) ListWorkflows(opts ListOptions) (*WorkflowList, error) {
	url := fmt.Sprintf("%s/automation/v4/flows", c.BaseURL)

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

	var result WorkflowList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse workflows response: %w", err)
	}

	return &result, nil
}

// GetWorkflow retrieves a single workflow by ID
func (c *Client) GetWorkflow(workflowID string) (*Workflow, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	url := fmt.Sprintf("%s/automation/v4/flows/%s", c.BaseURL, workflowID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Workflow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse workflow response: %w", err)
	}

	return &result, nil
}

// CreateWorkflow creates a new workflow
func (c *Client) CreateWorkflow(data map[string]interface{}) (*Workflow, error) {
	url := fmt.Sprintf("%s/automation/v4/flows", c.BaseURL)

	body, err := c.post(url, data)
	if err != nil {
		return nil, err
	}

	var result Workflow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse workflow response: %w", err)
	}

	return &result, nil
}

// UpdateWorkflow updates an existing workflow
func (c *Client) UpdateWorkflow(workflowID string, data map[string]interface{}) (*Workflow, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	url := fmt.Sprintf("%s/automation/v4/flows/%s", c.BaseURL, workflowID)

	body, err := c.patch(url, data)
	if err != nil {
		return nil, err
	}

	var result Workflow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse workflow response: %w", err)
	}

	return &result, nil
}

// DeleteWorkflow deletes a workflow by ID
func (c *Client) DeleteWorkflow(workflowID string) error {
	if workflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	url := fmt.Sprintf("%s/automation/v4/flows/%s", c.BaseURL, workflowID)

	_, err := c.delete(url)
	return err
}

// WorkflowEnrollment represents an enrollment in a workflow
type WorkflowEnrollment struct {
	ObjectID     string `json:"objectId"`
	ObjectType   string `json:"objectType,omitempty"`
	EnrolledAt   string `json:"enrolledAt,omitempty"`
	Status       string `json:"status,omitempty"`
	EnrollmentID string `json:"enrollmentId,omitempty"`
}

// WorkflowEnrollmentList represents a paginated list of enrollments
type WorkflowEnrollmentList struct {
	Results []WorkflowEnrollment `json:"results"`
	Paging  *Paging              `json:"paging,omitempty"`
}

// EnrollInWorkflow enrolls an object in a workflow
func (c *Client) EnrollInWorkflow(workflowID string, objectID string) error {
	if workflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}
	if objectID == "" {
		return fmt.Errorf("object ID is required")
	}

	url := fmt.Sprintf("%s/automation/v4/flows/%s/enrollments/start", c.BaseURL, workflowID)

	data := map[string]interface{}{
		"objectId": objectID,
	}

	_, err := c.post(url, data)
	return err
}

// ListWorkflowEnrollments lists enrollments for a workflow
func (c *Client) ListWorkflowEnrollments(workflowID string, opts ListOptions) (*WorkflowEnrollmentList, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	url := fmt.Sprintf("%s/automation/v4/flows/%s/enrollments", c.BaseURL, workflowID)

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

	var result WorkflowEnrollmentList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse enrollments response: %w", err)
	}

	return &result, nil
}
