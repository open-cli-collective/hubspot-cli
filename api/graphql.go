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

// GraphQL Introspection types

// IntrospectionType represents a GraphQL type from schema introspection
type IntrospectionType struct {
	Kind          string                 `json:"kind"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Fields        []IntrospectionField   `json:"fields,omitempty"`
	InputFields   []IntrospectionField   `json:"inputFields,omitempty"`
	EnumValues    []IntrospectionEnumVal `json:"enumValues,omitempty"`
	Interfaces    []IntrospectionTypeRef `json:"interfaces,omitempty"`
	PossibleTypes []IntrospectionTypeRef `json:"possibleTypes,omitempty"`
}

// IntrospectionField represents a field in a GraphQL type
type IntrospectionField struct {
	Name              string               `json:"name"`
	Description       string               `json:"description,omitempty"`
	Type              IntrospectionTypeRef `json:"type"`
	Args              []IntrospectionArg   `json:"args,omitempty"`
	IsDeprecated      bool                 `json:"isDeprecated,omitempty"`
	DeprecationReason string               `json:"deprecationReason,omitempty"`
}

// IntrospectionArg represents an argument on a GraphQL field
type IntrospectionArg struct {
	Name         string               `json:"name"`
	Description  string               `json:"description,omitempty"`
	Type         IntrospectionTypeRef `json:"type"`
	DefaultValue *string              `json:"defaultValue,omitempty"`
}

// IntrospectionTypeRef represents a type reference (can be nested for non-null/list)
type IntrospectionTypeRef struct {
	Kind   string                `json:"kind"`
	Name   *string               `json:"name,omitempty"`
	OfType *IntrospectionTypeRef `json:"ofType,omitempty"`
}

// IntrospectionEnumVal represents an enum value
type IntrospectionEnumVal struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	IsDeprecated      bool   `json:"isDeprecated,omitempty"`
	DeprecationReason string `json:"deprecationReason,omitempty"`
}

// IntrospectionSchema represents the full schema from introspection
type IntrospectionSchema struct {
	QueryType        *IntrospectionTypeRef `json:"queryType"`
	MutationType     *IntrospectionTypeRef `json:"mutationType,omitempty"`
	SubscriptionType *IntrospectionTypeRef `json:"subscriptionType,omitempty"`
	Types            []IntrospectionType   `json:"types"`
}

// IntrospectionResponse wraps the schema introspection response
type IntrospectionResponse struct {
	Schema IntrospectionSchema `json:"__schema"`
}

// TypeName returns the full type name including wrappers (NonNull, List)
func (t IntrospectionTypeRef) TypeName() string {
	if t.Name != nil {
		return *t.Name
	}
	if t.OfType != nil {
		inner := t.OfType.TypeName()
		switch t.Kind {
		case "NON_NULL":
			return inner + "!"
		case "LIST":
			return "[" + inner + "]"
		}
		return inner
	}
	return ""
}

// IntrospectSchema fetches the GraphQL schema via introspection
func (c *Client) IntrospectSchema() (*IntrospectionSchema, error) {
	query := `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      kind
      name
      description
      fields(includeDeprecated: true) {
        name
        description
        args {
          name
          description
          type {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
                ofType {
                  kind
                  name
                }
              }
            }
          }
          defaultValue
        }
        type {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
        isDeprecated
        deprecationReason
      }
      inputFields {
        name
        description
        type {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
            }
          }
        }
        defaultValue
      }
      interfaces {
        kind
        name
      }
      enumValues(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
      }
      possibleTypes {
        kind
        name
      }
    }
  }
}`

	resp, err := c.ExecuteGraphQL(query, nil)
	if err != nil {
		return nil, err
	}

	if resp.HasErrors() {
		return nil, fmt.Errorf("introspection failed: %s", resp.ErrorMessages())
	}

	var result IntrospectionResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse introspection response: %w", err)
	}

	return &result.Schema, nil
}

// GetType retrieves a specific type from the schema by name
func (s *IntrospectionSchema) GetType(name string) *IntrospectionType {
	for i := range s.Types {
		if s.Types[i].Name == name {
			return &s.Types[i]
		}
	}
	return nil
}

// GetRootTypes returns the main queryable types (non-internal types)
func (s *IntrospectionSchema) GetRootTypes() []IntrospectionType {
	var result []IntrospectionType
	for _, t := range s.Types {
		// Skip internal types (start with __)
		if len(t.Name) > 0 && t.Name[0:1] != "_" {
			// Include OBJECT and INTERFACE types that have fields
			if (t.Kind == "OBJECT" || t.Kind == "INTERFACE") && len(t.Fields) > 0 {
				result = append(result, t)
			}
		}
	}
	return result
}
