package view

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

// Format represents the output format
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatPlain Format = "plain"
)

// View handles output formatting.
//
// Status, progress, and banner text are written to Err (stderr) so that Out
// (stdout) carries only the primary rendered result. This keeps `--output json`
// output valid and parseable by tools like jq (issue #52): in JSON mode stdout
// is only the JSON payload, and in human/plain mode only the rendered result
// lands on stdout while status chatter goes to stderr.
type View struct {
	Format  Format
	NoColor bool
	Out     io.Writer
	Err     io.Writer
}

// New creates a new View with the given format
func New(format string, noColor bool) *View {
	v := &View{
		Format:  Format(format),
		NoColor: noColor,
		Out:     os.Stdout,
		Err:     os.Stderr,
	}

	if noColor {
		color.NoColor = true
	}

	return v
}

// Table renders data as a formatted table
func (v *View) Table(headers []string, rows [][]string) error {
	if v.Format == FormatJSON {
		return fmt.Errorf("use JSON method for JSON output")
	}

	if v.Format == FormatPlain {
		return v.Plain(rows)
	}

	w := tabwriter.NewWriter(v.Out, 0, 0, 2, ' ', 0)

	// Print headers
	headerLine := strings.Join(headers, "\t")
	fmt.Fprintln(w, color.New(color.Bold).Sprint(headerLine))

	// Print rows
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}

// JSON renders data as JSON
func (v *View) JSON(data interface{}) error {
	enc := json.NewEncoder(v.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// Plain renders rows as tab-separated values without headers
func (v *View) Plain(rows [][]string) error {
	for _, row := range rows {
		fmt.Fprintln(v.Out, strings.Join(row, "\t"))
	}
	return nil
}

// Render renders data based on the current format
func (v *View) Render(headers []string, rows [][]string, jsonData interface{}) error {
	switch v.Format {
	case FormatJSON:
		return v.JSON(jsonData)
	case FormatPlain:
		return v.Plain(rows)
	default:
		return v.Table(headers, rows)
	}
}

// Success prints a success status message to stderr.
func (v *View) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(v.Err, color.GreenString("✓ %s", msg))
}

// Error prints an error message to stderr.
func (v *View) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(v.Err, color.RedString("✗ %s", msg))
}

// Warning prints a warning message to stderr.
func (v *View) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(v.Err, color.YellowString("⚠ %s", msg))
}

// Info prints an informational status message to stderr.
func (v *View) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(v.Err, msg)
}

// PrintStatus prints an unformatted status message to stderr (no trailing newline).
func (v *View) PrintStatus(format string, args ...interface{}) {
	fmt.Fprintf(v.Err, format, args...)
}

// PrintlnStatus prints a status message to stderr with a trailing newline.
func (v *View) PrintlnStatus(format string, args ...interface{}) {
	fmt.Fprintln(v.Err, fmt.Sprintf(format, args...))
}
