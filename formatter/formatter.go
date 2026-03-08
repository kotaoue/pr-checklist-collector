package formatter

// Check represents a single checklist item.
type Check struct {
	Name string `json:"name"`
	Done bool   `json:"done"`
}

// Formatter is the interface for serializing checks into a specific file format.
// Implementing this interface allows future support for additional formats (e.g. YAML, CSV).
type Formatter interface {
	// Format serializes the given checks into bytes in the formatter's target format.
	Format(checks []Check) ([]byte, error)
	// Extension returns the file extension for this format (e.g. "json").
	Extension() string
}
