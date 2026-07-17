package shared

import (
	"reflect"
	"testing"

	"github.com/open-cli-collective/hubspot-cli/api"
)

func TestParseFilters(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []api.SearchFilter
		wantErr bool
	}{
		{
			name:  "shorthand EQ",
			input: []string{"hs_task_status=NOT_STARTED"},
			want: []api.SearchFilter{
				{PropertyName: "hs_task_status", Operator: "EQ", Value: "NOT_STARTED"},
			},
		},
		{
			name:  "shorthand NEQ",
			input: []string{"hs_email_status!=BOUNCED"},
			want: []api.SearchFilter{
				{PropertyName: "hs_email_status", Operator: "NEQ", Value: "BOUNCED"},
			},
		},
		{
			name:  "shorthand GTE",
			input: []string{"hs_task_priority>=2"},
			want: []api.SearchFilter{
				{PropertyName: "hs_task_priority", Operator: "GTE", Value: "2"},
			},
		},
		{
			name:  "shorthand LTE does not split on equals",
			input: []string{"hubspot_owner_id<=99"},
			want: []api.SearchFilter{
				{PropertyName: "hubspot_owner_id", Operator: "LTE", Value: "99"},
			},
		},
		{
			name:  "shorthand GT",
			input: []string{"amount>100"},
			want: []api.SearchFilter{
				{PropertyName: "amount", Operator: "GT", Value: "100"},
			},
		},
		{
			name:  "shorthand LT",
			input: []string{"amount<100"},
			want: []api.SearchFilter{
				{PropertyName: "amount", Operator: "LT", Value: "100"},
			},
		},
		{
			name:  "explicit operator with value",
			input: []string{"hs_email_subject:CONTAINS_TOKEN:Dev Academy"},
			want: []api.SearchFilter{
				{PropertyName: "hs_email_subject", Operator: "CONTAINS_TOKEN", Value: "Dev Academy"},
			},
		},
		{
			name:  "explicit EQ operator",
			input: []string{"hs_task_status:EQ:NOT_STARTED"},
			want: []api.SearchFilter{
				{PropertyName: "hs_task_status", Operator: "EQ", Value: "NOT_STARTED"},
			},
		},
		{
			// Lowercase tokens after the first colon are treated as part of a
			// shorthand value, not as an explicit operator. With no shorthand
			// comparison token present this is an error, which keeps values that
			// happen to contain colons (URLs, times) from being misparsed.
			name:    "lowercase operator-looking token is not an explicit operator",
			input:   []string{"hs_email_subject:contains_token:hello"},
			wantErr: true,
		},
		{
			name:  "HAS_PROPERTY without value",
			input: []string{"hubspot_owner_id:HAS_PROPERTY"},
			want: []api.SearchFilter{
				{PropertyName: "hubspot_owner_id", Operator: "HAS_PROPERTY"},
			},
		},
		{
			name:  "BETWEEN range",
			input: []string{"hs_task_priority:BETWEEN:1:3"},
			want: []api.SearchFilter{
				{PropertyName: "hs_task_priority", Operator: "BETWEEN", Value: "1", HighValue: "3"},
			},
		},
		{
			name:  "IN set membership",
			input: []string{"hs_task_status:IN:NOT_STARTED,IN_PROGRESS"},
			want: []api.SearchFilter{
				{PropertyName: "hs_task_status", Operator: "IN", Values: []string{"NOT_STARTED", "IN_PROGRESS"}},
			},
		},
		{
			name:  "NOT_IN set membership",
			input: []string{"hs_task_status:NOT_IN:COMPLETED,DEFERRED"},
			want: []api.SearchFilter{
				{PropertyName: "hs_task_status", Operator: "NOT_IN", Values: []string{"COMPLETED", "DEFERRED"}},
			},
		},
		{
			name:  "ISO date converted to unix millis for date property (shorthand)",
			input: []string{"hs_timestamp<=2026-03-17"},
			want: []api.SearchFilter{
				// 2026-03-17T00:00:00Z == 1773705600000 ms
				{PropertyName: "hs_timestamp", Operator: "LTE", Value: "1773705600000"},
			},
		},
		{
			name:  "ISO datetime converted for date property",
			input: []string{"hs_timestamp=2026-03-17T12:00:00Z"},
			want: []api.SearchFilter{
				{PropertyName: "hs_timestamp", Operator: "EQ", Value: "1773748800000"},
			},
		},
		{
			name:  "BETWEEN with ISO dates converts both bounds",
			input: []string{"hs_timestamp:BETWEEN:2026-03-17:2026-03-18"},
			want: []api.SearchFilter{
				{PropertyName: "hs_timestamp", Operator: "BETWEEN", Value: "1773705600000", HighValue: "1773792000000"},
			},
		},
		{
			name:  "numeric value for date property is left unchanged",
			input: []string{"hs_timestamp>=1773705600000"},
			want: []api.SearchFilter{
				{PropertyName: "hs_timestamp", Operator: "GTE", Value: "1773705600000"},
			},
		},
		{
			name:  "date-looking value on non-date property is not converted",
			input: []string{"hs_email_subject=2026-03-17"},
			want: []api.SearchFilter{
				{PropertyName: "hs_email_subject", Operator: "EQ", Value: "2026-03-17"},
			},
		},
		{
			name:  "multiple filters",
			input: []string{"hs_task_status=NOT_STARTED", "hubspot_owner_id=77999105"},
			want: []api.SearchFilter{
				{PropertyName: "hs_task_status", Operator: "EQ", Value: "NOT_STARTED"},
				{PropertyName: "hubspot_owner_id", Operator: "EQ", Value: "77999105"},
			},
		},
		{
			name:    "empty filter is an error",
			input:   []string{""},
			wantErr: true,
		},
		{
			name:    "no operator is an error",
			input:   []string{"hs_task_status"},
			wantErr: true,
		},
		{
			name:    "missing property name is an error",
			input:   []string{"=NOT_STARTED"},
			wantErr: true,
		},
		{
			name:    "BETWEEN with one value is an error",
			input:   []string{"hs_task_priority:BETWEEN:1"},
			wantErr: true,
		},
		{
			name:    "value operator without value is an error",
			input:   []string{"hs_email_subject:CONTAINS_TOKEN"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilters(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFilters(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFilters(%v) unexpected error: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFilters(%v) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFiltersEmptyInput(t *testing.T) {
	got, err := ParseFilters(nil)
	if err != nil {
		t.Fatalf("ParseFilters(nil) unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("ParseFilters(nil) = %+v, want nil", got)
	}
}

func TestParseSort(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []api.SearchSort
		wantErr bool
	}{
		{
			name:  "ascending explicit",
			input: []string{"hs_timestamp:asc"},
			want:  []api.SearchSort{{PropertyName: "hs_timestamp", Direction: "ASCENDING"}},
		},
		{
			name:  "descending explicit",
			input: []string{"hs_timestamp:desc"},
			want:  []api.SearchSort{{PropertyName: "hs_timestamp", Direction: "DESCENDING"}},
		},
		{
			name:  "default direction is ascending",
			input: []string{"hs_timestamp"},
			want:  []api.SearchSort{{PropertyName: "hs_timestamp", Direction: "ASCENDING"}},
		},
		{
			name:  "long form directions",
			input: []string{"createdate:descending", "amount:ascending"},
			want: []api.SearchSort{
				{PropertyName: "createdate", Direction: "DESCENDING"},
				{PropertyName: "amount", Direction: "ASCENDING"},
			},
		},
		{
			name:  "case-insensitive direction",
			input: []string{"hs_timestamp:DESC"},
			want:  []api.SearchSort{{PropertyName: "hs_timestamp", Direction: "DESCENDING"}},
		},
		{
			name:    "invalid direction is an error",
			input:   []string{"hs_timestamp:sideways"},
			wantErr: true,
		},
		{
			name:    "empty sort is an error",
			input:   []string{""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSort(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSort(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSort(%v) unexpected error: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSort(%v) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
