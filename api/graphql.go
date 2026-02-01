package api

import (
	"encoding/json"
	"fmt"
)

// GraphQLRequest represents a GraphQL query request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL query response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLLocation      `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLLocation represents a location in a GraphQL query where an error occurred
type GraphQLLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Error returns the error message
func (e GraphQLError) Error() string {
	return e.Message
}

// HasErrors returns true if the response contains errors
func (r *GraphQLResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorMessages returns all error messages as a single string
func (r *GraphQLResponse) ErrorMessages() string {
	if !r.HasErrors() {
		return ""
	}
	if len(r.Errors) == 1 {
		return r.Errors[0].Message
	}
	msg := ""
	for i, e := range r.Errors {
		if i > 0 {
			msg += "; "
		}
		msg += e.Message
	}
	return msg
}

// ExecuteGraphQL executes a GraphQL query
func (c *Client) ExecuteGraphQL(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	url := fmt.Sprintf("%s/collector/graphql", c.BaseURL)

	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	respBytes, err := c.post(url, bodyBytes)
	if err != nil {
		return nil, err
	}

	var result GraphQLResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return &result, nil
}
