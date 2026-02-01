package api

import (
	"encoding/json"
	"fmt"
)

// Pipeline represents a HubSpot pipeline
type Pipeline struct {
	ID           string          `json:"id"`
	Label        string          `json:"label"`
	DisplayOrder int             `json:"displayOrder"`
	Archived     bool            `json:"archived,omitempty"`
	Stages       []PipelineStage `json:"stages,omitempty"`
	CreatedAt    string          `json:"createdAt,omitempty"`
	UpdatedAt    string          `json:"updatedAt,omitempty"`
	ArchivedAt   string          `json:"archivedAt,omitempty"`
}

// PipelineStage represents a stage within a pipeline
type PipelineStage struct {
	ID           string                 `json:"id"`
	Label        string                 `json:"label"`
	DisplayOrder int                    `json:"displayOrder"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Archived     bool                   `json:"archived,omitempty"`
	CreatedAt    string                 `json:"createdAt,omitempty"`
	UpdatedAt    string                 `json:"updatedAt,omitempty"`
	ArchivedAt   string                 `json:"archivedAt,omitempty"`
}

// PipelineList represents a list of pipelines
type PipelineList struct {
	Results []Pipeline `json:"results"`
}

// ListPipelines lists all pipelines for an object type
func (c *Client) ListPipelines(objectType ObjectType) (*PipelineList, error) {
	url := fmt.Sprintf("%s/crm/v3/pipelines/%s", c.BaseURL, objectType)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result PipelineList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetPipeline retrieves a specific pipeline by ID
func (c *Client) GetPipeline(objectType ObjectType, pipelineID string) (*Pipeline, error) {
	if pipelineID == "" {
		return nil, fmt.Errorf("pipeline ID is required")
	}

	url := fmt.Sprintf("%s/crm/v3/pipelines/%s/%s", c.BaseURL, objectType, pipelineID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result Pipeline
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetPipelineStages retrieves all stages for a pipeline
func (c *Client) GetPipelineStages(objectType ObjectType, pipelineID string) ([]PipelineStage, error) {
	if pipelineID == "" {
		return nil, fmt.Errorf("pipeline ID is required")
	}

	url := fmt.Sprintf("%s/crm/v3/pipelines/%s/%s/stages", c.BaseURL, objectType, pipelineID)

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []PipelineStage `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Results, nil
}
