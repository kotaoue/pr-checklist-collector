package formatter

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJSONFormatter_Extension(t *testing.T) {
	f := &JSONFormatter{}
	if got := f.Extension(); got != "json" {
		t.Errorf("Extension() = %q, want %q", got, "json")
	}
}

func TestJSONFormatter_Format(t *testing.T) {
	checks := []Check{
		{Name: "dog", Done: true},
		{Name: "cat", Done: false},
	}

	f := &JSONFormatter{}
	data, err := f.Format("2026-03-08", checks)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Unmarshal and compare to avoid map-ordering sensitivity.
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got["date"] != "2026-03-08" {
		t.Errorf("date = %v, want %q", got["date"], "2026-03-08")
	}
	checksMap, ok := got["checks"].(map[string]interface{})
	if !ok {
		t.Fatalf("checks is not a map: %T", got["checks"])
	}
	if checksMap["dog"] != true {
		t.Errorf("checks[dog] = %v, want true", checksMap["dog"])
	}
	if checksMap["cat"] != false {
		t.Errorf("checks[cat] = %v, want false", checksMap["cat"])
	}
}

func TestJSONFormatter_Format_Empty(t *testing.T) {
	f := &JSONFormatter{}
	data, err := f.Format("2026-03-08", []Check{})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got["date"] != "2026-03-08" {
		t.Errorf("date = %v, want %q", got["date"], "2026-03-08")
	}
	checksMap, ok := got["checks"].(map[string]interface{})
	if !ok {
		t.Fatalf("checks is not a map: %T", got["checks"])
	}
	if len(checksMap) != 0 {
		t.Errorf("checks map len = %d, want 0", len(checksMap))
	}
}

func TestCheck_Fields(t *testing.T) {
	c := Check{Name: "run", Done: true}
	if c.Name != "run" || !c.Done {
		t.Errorf("unexpected Check fields: %+v", c)
	}
}

func TestFormatterInterface(t *testing.T) {
	// Verify JSONFormatter satisfies the Formatter interface at compile time.
	var _ Formatter = (*JSONFormatter)(nil)
	_ = reflect.TypeOf((*Formatter)(nil)).Elem()
}

func TestNewFromPath_JSON(t *testing.T) {
	f, err := NewFromPath("results/results.json")
	if err != nil {
		t.Fatalf("NewFromPath() error = %v", err)
	}
	if f.Extension() != "json" {
		t.Errorf("Extension() = %q, want %q", f.Extension(), "json")
	}
}

func TestNewFromPath_Unsupported(t *testing.T) {
	_, err := NewFromPath("output.yaml")
	if err == nil {
		t.Error("NewFromPath() expected error for unsupported format, got nil")
	}
}
