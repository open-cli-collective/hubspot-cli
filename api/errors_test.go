package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name   string
		apiErr *APIError
		want   string
	}{
		{
			name: "with message only",
			apiErr: &APIError{
				StatusCode: 400,
				Message:    "The request was invalid",
			},
			want: "The request was invalid",
		},
		{
			name: "with message and error type",
			apiErr: &APIError{
				StatusCode: 400,
				Message:    "Invalid property",
				ErrorType:  "VALIDATION_ERROR",
			},
			want: "Invalid property; type: VALIDATION_ERROR",
		},
		{
			name: "with all fields",
			apiErr: &APIError{
				StatusCode:    429,
				Message:       "You have reached your daily limit.",
				ErrorType:     "RATE_LIMIT",
				CorrelationID: "c033cdaa-2c40-4a64-ae48-b4cec88dad24",
			},
			want: "You have reached your daily limit.; type: RATE_LIMIT; correlationId: c033cdaa-2c40-4a64-ae48-b4cec88dad24",
		},
		{
			name: "empty - just status",
			apiErr: &APIError{
				StatusCode: 500,
			},
			want: "API error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiErr.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
		wantMsg    string
	}{
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{}`,
			wantErr:    ErrUnauthorized,
		},
		{
			name:       "401 with message",
			statusCode: http.StatusUnauthorized,
			body:       `{"status": "error", "message": "The access token is invalid"}`,
			wantErr:    ErrUnauthorized,
			wantMsg:    "The access token is invalid",
		},
		{
			name:       "403 forbidden",
			statusCode: http.StatusForbidden,
			body:       `{}`,
			wantErr:    ErrForbidden,
		},
		{
			name:       "403 with message",
			statusCode: http.StatusForbidden,
			body:       `{"status": "error", "message": "Missing required scope: crm.objects.contacts.read"}`,
			wantErr:    ErrForbidden,
			wantMsg:    "Missing required scope",
		},
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			body:       `{}`,
			wantErr:    ErrNotFound,
		},
		{
			name:       "404 with message",
			statusCode: http.StatusNotFound,
			body:       `{"status": "error", "message": "Contact not found"}`,
			wantErr:    ErrNotFound,
			wantMsg:    "Contact not found",
		},
		{
			name:       "400 bad request",
			statusCode: http.StatusBadRequest,
			body:       `{}`,
			wantErr:    ErrBadRequest,
		},
		{
			name:       "400 with message",
			statusCode: http.StatusBadRequest,
			body:       `{"status": "error", "message": "Property 'email' is required"}`,
			wantErr:    ErrBadRequest,
			wantMsg:    "Property 'email' is required",
		},
		{
			name:       "429 rate limited",
			statusCode: http.StatusTooManyRequests,
			body:       `{"status": "error", "message": "You have reached your daily limit.", "errorType": "RATE_LIMIT"}`,
			wantErr:    ErrRateLimited,
		},
		{
			name:       "500 server error",
			statusCode: http.StatusInternalServerError,
			body:       `{}`,
			wantErr:    ErrServerError,
		},
		{
			name:       "500 with message",
			statusCode: http.StatusInternalServerError,
			body:       `{"status": "error", "message": "Internal server error"}`,
			wantErr:    ErrServerError,
			wantMsg:    "Internal server error",
		},
		{
			name:       "502 bad gateway",
			statusCode: http.StatusBadGateway,
			body:       `{}`,
			wantErr:    ErrServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			rec.WriteHeader(tt.statusCode)
			rec.WriteString(tt.body)
			resp := rec.Result()

			err := ParseAPIError(resp, []byte(tt.body))
			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)

			if tt.wantMsg != "" {
				assert.Contains(t, err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestParseAPIError_418_NonStandard(t *testing.T) {
	// Test a non-standard status code that isn't explicitly handled
	rec := httptest.NewRecorder()
	rec.WriteHeader(418) // I'm a teapot
	body := `{"status": "error", "message": "I'm a teapot", "errorType": "TEAPOT"}`
	rec.WriteString(body)
	resp := rec.Result()

	err := ParseAPIError(resp, []byte(body))

	// Should return an APIError, not a sentinel error
	var apiErr *APIError
	assert.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 418, apiErr.StatusCode)
	assert.Contains(t, err.Error(), "I'm a teapot")
}

func TestIsNotFound(t *testing.T) {
	assert.True(t, IsNotFound(ErrNotFound))
	assert.True(t, IsNotFound(fmt.Errorf("wrapped: %w", ErrNotFound)))
	assert.False(t, IsNotFound(ErrUnauthorized))
	assert.False(t, IsNotFound(nil))
}

func TestIsUnauthorized(t *testing.T) {
	assert.True(t, IsUnauthorized(ErrUnauthorized))
	assert.True(t, IsUnauthorized(fmt.Errorf("wrapped: %w", ErrUnauthorized)))
	assert.False(t, IsUnauthorized(ErrNotFound))
	assert.False(t, IsUnauthorized(nil))
}

func TestIsForbidden(t *testing.T) {
	assert.True(t, IsForbidden(ErrForbidden))
	assert.True(t, IsForbidden(fmt.Errorf("wrapped: %w", ErrForbidden)))
	assert.False(t, IsForbidden(ErrNotFound))
	assert.False(t, IsForbidden(nil))
}

func TestIsRateLimited(t *testing.T) {
	assert.True(t, IsRateLimited(ErrRateLimited))
	assert.True(t, IsRateLimited(fmt.Errorf("wrapped: %w", ErrRateLimited)))
	assert.False(t, IsRateLimited(ErrNotFound))
	assert.False(t, IsRateLimited(nil))
}
