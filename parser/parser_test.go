package parser

import (
	"testing"

	"github.com/kotaoue/pr-checklist-collector/formatter"
)

func TestParseChecks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []formatter.Check
	}{
		{
			name:  "three items",
			input: "dog\ncat\nbird",
			want: []formatter.Check{
				{Name: "dog", Done: false},
				{Name: "cat", Done: false},
				{Name: "bird", Done: false},
			},
		},
		{
			name:  "trims whitespace",
			input: "  dog  \n  cat  ",
			want: []formatter.Check{
				{Name: "dog", Done: false},
				{Name: "cat", Done: false},
			},
		},
		{
			name:  "skips blank lines",
			input: "dog\n\ncat\n\n",
			want: []formatter.Check{
				{Name: "dog", Done: false},
				{Name: "cat", Done: false},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseChecks(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseChecks() len = %d, want %d", len(got), len(tt.want))
			}
			for i, c := range got {
				if c != tt.want[i] {
					t.Errorf("ParseChecks()[%d] = %+v, want %+v", i, c, tt.want[i])
				}
			}
		})
	}
}

func TestParseBody(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []formatter.Check
	}{
		{
			name:  "checked and unchecked items",
			input: "- [x] dog\n- [ ] cat\n- [x] bird",
			want: []formatter.Check{
				{Name: "dog", Done: true},
				{Name: "cat", Done: false},
				{Name: "bird", Done: true},
			},
		},
		{
			name:  "uppercase X is treated as checked",
			input: "- [X] dog\n- [ ] cat",
			want: []formatter.Check{
				{Name: "dog", Done: true},
				{Name: "cat", Done: false},
			},
		},
		{
			name:  "non-checkbox lines are ignored",
			input: "Some description\n- [x] task\nAnother line\n- [ ] task2",
			want: []formatter.Check{
				{Name: "task", Done: true},
				{Name: "task2", Done: false},
			},
		},
		{
			name:  "trims surrounding whitespace on lines",
			input: "  - [x] dog  \n  - [ ] cat  ",
			want: []formatter.Check{
				{Name: "dog", Done: true},
				{Name: "cat", Done: false},
			},
		},
		{
			name:  "empty body",
			input: "",
			want:  nil,
		},
		{
			name:  "all unchecked",
			input: "- [ ] a\n- [ ] b",
			want: []formatter.Check{
				{Name: "a", Done: false},
				{Name: "b", Done: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseBody(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseBody() len = %d, want %d", len(got), len(tt.want))
			}
			for i, c := range got {
				if c != tt.want[i] {
					t.Errorf("ParseBody()[%d] = %+v, want %+v", i, c, tt.want[i])
				}
			}
		})
	}
}
