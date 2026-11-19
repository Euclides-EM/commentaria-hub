package llm

import "strings"

// errorBody formats provider output for inclusion in an error message.
// Keep the complete body for now so provider diagnostics are not hidden.
func errorBody(body []byte) string {
	return strings.TrimSpace(string(body))
}
