package shared

import "testing"

func TestFormatBool(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		want  string
	}{
		{"true returns Yes", true, "Yes"},
		{"false returns No", false, "No"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBool(tt.input)
			if got != tt.want {
				t.Errorf("FormatBool(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
