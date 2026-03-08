package formatter

import "encoding/json"

// jsonResult is the top-level JSON structure written to the output file.
type jsonResult struct {
	Date   string          `json:"date"`
	Checks map[string]bool `json:"checks"`
}

// JSONFormatter formats checks as a JSON object.
type JSONFormatter struct{}

// Extension returns "json".
func (f *JSONFormatter) Extension() string { return "json" }

// Format marshals checks into an indented JSON object with a date field and a
// checks map (item name → done status), matching the FitnessStreak result format.
func (f *JSONFormatter) Format(date string, checks []Check) ([]byte, error) {
	result := jsonResult{
		Date:   date,
		Checks: make(map[string]bool, len(checks)),
	}
	for _, c := range checks {
		result.Checks[c.Name] = c.Done
	}
	return json.MarshalIndent(result, "", "  ")
}
