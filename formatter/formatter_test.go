package formatter

import (
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
		{Name: "dog", Done: false},
		{Name: "cat", Done: false},
		{Name: "bird", Done: false},
	}

	f := &JSONFormatter{}
	data, err := f.Format(checks)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := `[
  {
    "name": "dog",
    "done": false
  },
  {
    "name": "cat",
    "done": false
  },
  {
    "name": "bird",
    "done": false
  }
]`
	if got := string(data); got != want {
		t.Errorf("Format() =\n%s\nwant:\n%s", got, want)
	}
}

func TestJSONFormatter_Format_Empty(t *testing.T) {
	f := &JSONFormatter{}
	data, err := f.Format([]Check{})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if string(data) != "[]" {
		t.Errorf("Format() = %q, want %q", string(data), "[]")
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
