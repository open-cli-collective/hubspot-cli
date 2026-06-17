package view

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestView returns a View that writes to the provided buffers, with color
// disabled so assertions can match raw text without ANSI escape codes.
func newTestView(format string) (*View, *bytes.Buffer, *bytes.Buffer) {
	var out, errBuf bytes.Buffer
	v := New(format, true) // noColor=true
	v.Out = &out
	v.Err = &errBuf
	return v, &out, &errBuf
}

// TestStatusMessagesGoToStderr ensures human-facing status/progress messages
// are written to stderr, not stdout. This is what keeps `--output json`
// output valid (issue #52): only the JSON payload may land on stdout.
func TestStatusMessagesGoToStderr(t *testing.T) {
	tests := []struct {
		name string
		call func(v *View)
		want string
	}{
		{"Success", func(v *View) { v.Success("created with ID %s", "98765") }, "created with ID 98765"},
		{"Info", func(v *View) { v.Info("Found %d contact(s)", 3) }, "Found 3 contact(s)"},
		{"Print", func(v *View) { v.Print("progress %d%%", 50) }, "progress 50%"},
		{"Println", func(v *View) { v.Println("More results available") }, "More results available"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, out, errBuf := newTestView("json")
			tt.call(v)

			assert.Empty(t, out.String(), "status message must NOT be written to stdout")
			assert.Contains(t, errBuf.String(), tt.want, "status message must be written to stderr")
		})
	}
}

// TestErrorAndWarningGoToStderr documents that Error and Warning remain on
// stderr (they always have); this is now consistent with Success/Info/Print.
func TestErrorAndWarningGoToStderr(t *testing.T) {
	v, out, errBuf := newTestView("json")

	v.Error("boom %s", "x")
	v.Warning("careful %s", "y")

	assert.Empty(t, out.String())
	assert.Contains(t, errBuf.String(), "boom x")
	assert.Contains(t, errBuf.String(), "careful y")
}

// TestStdoutIsValidJSON simulates a command that emits a status message
// followed by a JSON payload, and verifies stdout alone is parseable JSON.
func TestStdoutIsValidJSON(t *testing.T) {
	v, out, _ := newTestView("json")

	// Order mirrors real commands: status first, then structured payload.
	v.Success("Note created with ID: %s", "98765")
	v.Info("Found 1 result")
	require.NoError(t, v.JSON(map[string]interface{}{
		"id":   "98765",
		"name": "Demo",
	}))
	v.Info("More results available. Use --after abc123 to get the next page.")

	stdout := out.String()
	assert.NotContains(t, stdout, "Note created", "no status banner may leak to stdout")
	assert.NotContains(t, stdout, "More results", "no pagination text may leak to stdout")

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &parsed),
		"stdout must be valid JSON, got: %q", stdout)
	assert.Equal(t, "98765", parsed["id"])
}

// TestJSONOutputUnaffectedByColorFlag is a small sanity check that JSON output
// is plain (no ANSI codes regardless of color setting).
func TestJSONOutputUnaffectedByColorFlag(t *testing.T) {
	v, out, _ := newTestView("json")
	require.NoError(t, v.JSON([]string{"a", "b"}))
	assert.False(t, strings.Contains(out.String(), "\x1b["), "JSON output must not contain ANSI escapes")
}
