package configcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "empty token",
			token:    "",
			expected: "",
		},
		{
			name:     "short token",
			token:    "abc",
			expected: "********",
		},
		{
			name:     "exactly 8 chars",
			token:    "12345678",
			expected: "********",
		},
		{
			name:     "9 chars",
			token:    "123456789",
			expected: "1234********6789",
		},
		{
			name:     "long token",
			token:    "abcd1234567890efghijklmnopqrstuv",
			expected: "abcd********stuv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}
