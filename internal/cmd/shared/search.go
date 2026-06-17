package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/open-cli-collective/hubspot-cli/api"
)

// shorthandOperators maps shorthand comparison tokens to HubSpot search
// operators. Longer tokens are listed first so prefix matching does not
// mistake ">=" for ">".
var shorthandOperators = []struct {
	token    string
	operator string
}{
	{">=", "GTE"},
	{"<=", "LTE"},
	{"!=", "NEQ"},
	{"=", "EQ"},
	{">", "GT"},
	{"<", "LT"},
}

// dateProperties are HubSpot properties whose values are stored as Unix
// millisecond timestamps. ISO-8601 dates supplied for these properties are
// converted automatically so users can write human-readable dates.
var dateProperties = map[string]bool{
	"hs_timestamp":               true,
	"hs_task_timestamp":          true,
	"hs_createdate":              true,
	"hs_lastmodifieddate":        true,
	"createdate":                 true,
	"lastmodifieddate":           true,
	"closedate":                  true,
	"hs_task_completion_date":    true,
	"hubspot_owner_assigneddate": true,
}

// ParseFilters converts raw --filter flag values into HubSpot search filters.
//
// Supported forms:
//
//	prop=value          EQ
//	prop!=value         NEQ
//	prop>=value         GTE
//	prop<=value         LTE
//	prop>value          GT
//	prop<value          LT
//	prop:OPERATOR       operators that take no value (e.g. HAS_PROPERTY)
//	prop:OPERATOR:value operators that take a value (e.g. CONTAINS_TOKEN, EQ)
//	prop:BETWEEN:lo:hi  range query (inclusive low and high values)
//	prop:IN:a,b,c       set membership (also NOT_IN)
//
// ISO-8601 date values for known date properties are converted to Unix
// milliseconds automatically.
func ParseFilters(raw []string) ([]api.SearchFilter, error) {
	var filters []api.SearchFilter

	for _, r := range raw {
		f, err := parseFilter(r)
		if err != nil {
			return nil, err
		}
		filters = append(filters, f)
	}

	return filters, nil
}

func parseFilter(raw string) (api.SearchFilter, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return api.SearchFilter{}, fmt.Errorf("empty filter")
	}

	// Explicit operator form uses a colon separator: prop:OPERATOR[:value...].
	// Operators must be written in uppercase (the HubSpot convention). We only
	// treat the segment after the first colon as an operator when it is all
	// uppercase letters/underscores, so shorthand values containing colons
	// (e.g. URLs or timestamps) are not misparsed as operators.
	if idx := strings.Index(trimmed, ":"); idx > 0 {
		prop := trimmed[:idx]
		rest := trimmed[idx+1:]
		opEnd := strings.Index(rest, ":")
		op := rest
		if opEnd >= 0 {
			op = rest[:opEnd]
		}
		if isOperatorToken(op) {
			return parseExplicitFilter(prop, strings.ToUpper(op), rest, opEnd)
		}
	}

	// Shorthand comparison form: prop<op>value.
	for _, sh := range shorthandOperators {
		if idx := strings.Index(trimmed, sh.token); idx > 0 {
			prop := strings.TrimSpace(trimmed[:idx])
			value := strings.TrimSpace(trimmed[idx+len(sh.token):])
			if prop == "" {
				return api.SearchFilter{}, fmt.Errorf("filter %q is missing a property name", raw)
			}
			return api.SearchFilter{
				PropertyName: prop,
				Operator:     sh.operator,
				Value:        convertDateValue(prop, value),
			}, nil
		}
	}

	return api.SearchFilter{}, fmt.Errorf("invalid filter %q: expected prop=value, prop>=value, or prop:OPERATOR:value", raw)
}

func parseExplicitFilter(prop, operator, rest string, opEnd int) (api.SearchFilter, error) {
	if prop == "" {
		return api.SearchFilter{}, fmt.Errorf("filter is missing a property name")
	}

	// Operators that take no value.
	if opEnd < 0 {
		switch operator {
		case "HAS_PROPERTY", "NOT_HAS_PROPERTY":
			return api.SearchFilter{PropertyName: prop, Operator: operator}, nil
		default:
			return api.SearchFilter{}, fmt.Errorf("operator %s requires a value (use %s:%s:value)", operator, prop, operator)
		}
	}

	value := rest[opEnd+1:]

	switch operator {
	case "BETWEEN":
		parts := strings.SplitN(value, ":", 2)
		if len(parts) != 2 {
			return api.SearchFilter{}, fmt.Errorf("BETWEEN requires two values (use %s:BETWEEN:low:high)", prop)
		}
		return api.SearchFilter{
			PropertyName: prop,
			Operator:     operator,
			Value:        convertDateValue(prop, parts[0]),
			HighValue:    convertDateValue(prop, parts[1]),
		}, nil
	case "IN", "NOT_IN":
		var values []string
		for _, v := range strings.Split(value, ",") {
			v = strings.TrimSpace(v)
			if v != "" {
				values = append(values, convertDateValue(prop, v))
			}
		}
		if len(values) == 0 {
			return api.SearchFilter{}, fmt.Errorf("%s requires at least one value (use %s:%s:a,b,c)", operator, prop, operator)
		}
		return api.SearchFilter{PropertyName: prop, Operator: operator, Values: values}, nil
	default:
		return api.SearchFilter{
			PropertyName: prop,
			Operator:     operator,
			Value:        convertDateValue(prop, value),
		}, nil
	}
}

// isOperatorToken reports whether s looks like a HubSpot operator name
// (uppercase letters and underscores, at least two characters). This keeps
// the explicit-form detection from swallowing shorthand values.
func isOperatorToken(s string) bool {
	if len(s) < 2 {
		return false
	}
	for _, r := range s {
		if (r < 'A' || r > 'Z') && r != '_' {
			return false
		}
	}
	return true
}

// convertDateValue converts an ISO-8601 date/datetime to a Unix millisecond
// string when the property is a known date property. Non-date properties and
// values that are already numeric or unparseable are returned unchanged.
func convertDateValue(prop, value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !dateProperties[prop] {
		return value
	}

	// Already a Unix millisecond timestamp.
	if isAllDigits(value) {
		return value
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return fmt.Sprintf("%d", t.UTC().UnixMilli())
		}
	}

	return value
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ParseSort converts raw --sort flag values into HubSpot search sorts.
//
// Supported forms:
//
//	prop          ascending (default)
//	prop:asc      ascending
//	prop:desc     descending
func ParseSort(raw []string) ([]api.SearchSort, error) {
	var sorts []api.SearchSort

	for _, r := range raw {
		trimmed := strings.TrimSpace(r)
		if trimmed == "" {
			return nil, fmt.Errorf("empty sort")
		}

		prop := trimmed
		direction := "ASCENDING"

		if idx := strings.LastIndex(trimmed, ":"); idx > 0 {
			prop = strings.TrimSpace(trimmed[:idx])
			dir := strings.ToLower(strings.TrimSpace(trimmed[idx+1:]))
			switch dir {
			case "asc", "ascending":
				direction = "ASCENDING"
			case "desc", "descending":
				direction = "DESCENDING"
			default:
				return nil, fmt.Errorf("invalid sort direction %q in %q (use asc or desc)", dir, r)
			}
		}

		if prop == "" {
			return nil, fmt.Errorf("sort %q is missing a property name", r)
		}

		sorts = append(sorts, api.SearchSort{PropertyName: prop, Direction: direction})
	}

	return sorts, nil
}
