package shared

// FormatBool returns "Yes" or "No" for boolean values.
// Used for human-readable output in table views.
func FormatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
