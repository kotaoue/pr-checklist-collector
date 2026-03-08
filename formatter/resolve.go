package formatter

import (
	"fmt"
	"strings"
)

// NewFromPath returns the Formatter appropriate for the given file path,
// determined by the file extension.
func NewFromPath(path string) (Formatter, error) {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".json"):
		return &JSONFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported file format for %q (supported: .json)", path)
	}
}
