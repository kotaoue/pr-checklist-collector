package formatter

import (
	"bytes"
	"encoding/json"
	"strings"
)

// JSONFormatter formats checks as a JSON object.
type JSONFormatter struct{}

// Extension returns "json".
func (f *JSONFormatter) Extension() string { return "json" }

// Format marshals checks into an indented JSON object with a date field and a
// named checks map (item name → done status), matching the FitnessStreak result format.
// checksKey is the name of the object key that holds the checklist map (e.g. "checks" or "exercises").
func (f *JSONFormatter) Format(date string, checksKey string, checks []Check) ([]byte, error) {
	checksMap := make(map[string]bool, len(checks))
	for _, c := range checks {
		checksMap[c.Name] = c.Done
	}

	// Marshal the inner map with standard 2-space indentation.
	innerData, err := json.MarshalIndent(checksMap, "", "  ")
	if err != nil {
		return nil, err
	}
	// Indent every line after the first by 2 spaces so the map aligns correctly
	// when embedded inside the outer object.
	indented := strings.ReplaceAll(string(innerData), "\n", "\n  ")

	// Build outer object manually so that "date" always appears first and the
	// checks key name is dynamic.
	dateData, err := json.Marshal(date)
	if err != nil {
		return nil, err
	}
	keyData, err := json.Marshal(checksKey)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("{\n  \"date\": ")
	buf.Write(dateData)
	buf.WriteString(",\n  ")
	buf.Write(keyData)
	buf.WriteString(": ")
	buf.WriteString(indented)
	buf.WriteString("\n}")
	return buf.Bytes(), nil
}
