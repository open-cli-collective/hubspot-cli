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
		{"PrintStatus", func(v *View) { v.PrintStatus("progress %d%%", 50) }, "progress 50%"},
		{"PrintlnStatus", func(v *View) { v.PrintlnStatus("More results available") }, "More results available"},
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
// stderr (they always have); this is now consistent with the status methods
// Success/Info/PrintStatus/PrintlnStatus.
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

// TestPrimaryOutputGoesToStdout guarantees the other half of the invariant:
// the primary rendered result (Table/Plain/Render/JSON) always lands on stdout,
// never stderr, even while status chatter is routed to stderr. If a "status"
// method ever leaked primary output to stderr, this would catch it.
func TestPrimaryOutputGoesToStdout(t *testing.T) {
	t.Run("table renders to stdout in human mode", func(t *testing.T) {
		v, out, errBuf := newTestView("table")
		// Interleave status messages with the primary rendered result.
		v.Info("Found 1 result")
		require.NoError(t, v.Render([]string{"ID", "NAME"}, [][]string{{"98765", "Demo"}}, nil))
		v.PrintlnStatus("More results available")

		stdout := out.String()
		assert.Contains(t, stdout, "98765", "primary rendered data must be on stdout")
		assert.Contains(t, stdout, "Demo", "primary rendered data must be on stdout")
		// Status chatter must not leak into the primary stdout stream.
		assert.NotContains(t, stdout, "Found 1 result")
		assert.NotContains(t, stdout, "More results available")
		assert.Contains(t, errBuf.String(), "Found 1 result")
		assert.Contains(t, errBuf.String(), "More results available")
	})

	t.Run("plain renders to stdout in plain mode", func(t *testing.T) {
		v, out, errBuf := newTestView("plain")
		v.Info("Found 1 result")
		require.NoError(t, v.Render(nil, [][]string{{"98765", "Demo"}}, nil))

		assert.Contains(t, out.String(), "98765", "primary rendered data must be on stdout")
		assert.NotContains(t, out.String(), "Found 1 result")
		assert.Contains(t, errBuf.String(), "Found 1 result")
	})
}

// TestJSONOutputUnaffectedByColorFlag is a small sanity check that JSON output
// is plain (no ANSI codes regardless of color setting).
func TestJSONOutputUnaffectedByColorFlag(t *testing.T) {
	v, out, _ := newTestView("json")
	require.NoError(t, v.JSON([]string{"a", "b"}))
	assert.False(t, strings.Contains(out.String(), "\x1b["), "JSON output must not contain ANSI escapes")
}
