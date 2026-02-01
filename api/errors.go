package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel errors
var (
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized: check your access token")
	ErrForbidden           = errors.New("forbidden: missing required scopes")
	ErrBadRequest          = errors.New("bad request")
	ErrRateLimited         = errors.New("rate limited: too many requests")
	ErrServerError         = errors.New("server error")
	ErrAccessTokenRequired = errors.New("access token is required")
)

// APIError represents an error response from the HubSpot API
type APIError struct {
	StatusCode    int
	Status        string `json:"status"`
	Message       string `json:"message"`
	ErrorType     string `json:"errorType"`
	CorrelationID string `json:"correlationId"`
}

func (e *APIError) Error() string {
	var parts []string

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.ErrorType != "" {
		parts = append(parts, fmt.Sprintf("type: %s", e.ErrorType))
	}

	if e.CorrelationID != "" {
		parts = append(parts, fmt.Sprintf("correlationId: %s", e.CorrelationID))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("API error (status %d)", e.StatusCode)
	}

	return strings.Join(parts, "; ")
}

// ParseAPIError parses an error response from the HubSpot API
func ParseAPIError(resp *http.Response, body []byte) error {
	apiErr := &APIError{StatusCode: resp.StatusCode}

	if len(body) > 0 {
		_ = json.Unmarshal(body, apiErr)
	}

	// Return sentinel errors for common status codes
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		if apiErr.Message != "" {
			return fmt.Errorf("%w: %s", ErrUnauthorized, apiErr.Message)
		}
		return ErrUnauthorized
	case http.StatusForbidden:
		if apiErr.Message != "" {
			return fmt.Errorf("%w: %s", ErrForbidden, apiErr.Message)
		}
		return ErrForbidden
	case http.StatusNotFound:
		if apiErr.Message != "" {
			return fmt.Errorf("%w: %s", ErrNotFound, apiErr.Message)
		}
		return ErrNotFound
	case http.StatusBadRequest:
		if apiErr.Message != "" {
			return fmt.Errorf("%w: %s", ErrBadRequest, apiErr.Message)
		}
		return ErrBadRequest
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		if resp.StatusCode >= 500 {
			if apiErr.Message != "" {
				return fmt.Errorf("%w: %s", ErrServerError, apiErr.Message)
			}
			return ErrServerError
		}
		return apiErr
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsRateLimited checks if an error is a rate limited error
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}
