package formatter

import "encoding/json"

// JSONFormatter formats checks as a JSON array.
type JSONFormatter struct{}

// Extension returns "json".
func (f *JSONFormatter) Extension() string { return "json" }

// Format marshals checks to indented JSON.
func (f *JSONFormatter) Format(checks []Check) ([]byte, error) {
	return json.MarshalIndent(checks, "", "  ")
}
