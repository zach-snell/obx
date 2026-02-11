package vault

import (
	"strings"
	"testing"
)

func TestExtractInlineFields(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []InlineField
	}{
		{
			name:    "standard field",
			content: "Some text\nstatus:: done\nMore text",
			expected: []InlineField{
				{Key: "status", Value: "done", Line: 2},
			},
		},
		{
			name:    "bracket field",
			content: "Text with [priority:: high] inline",
			expected: []InlineField{
				{Key: "priority", Value: "high", Line: 1},
			},
		},
		{
			name:    "paren field",
			content: "Text with (due:: 2024-01-15) hidden key",
			expected: []InlineField{
				{Key: "due", Value: "2024-01-15", Line: 1},
			},
		},
		{
			name:    "multiple fields",
			content: "status:: active\npriority:: high\ndue:: tomorrow",
			expected: []InlineField{
				{Key: "status", Value: "active", Line: 1},
				{Key: "priority", Value: "high", Line: 2},
				{Key: "due", Value: "tomorrow", Line: 3},
			},
		},
		{
			name:    "field with spaces",
			content: "author :: John Doe",
			expected: []InlineField{
				{Key: "author", Value: "John Doe", Line: 1},
			},
		},
		{
			name:     "no fields",
			content:  "Just regular text\nNo fields here",
			expected: nil,
		},
		{
			name:    "field with link",
			content: "project:: [[My Project]]",
			expected: []InlineField{
				{Key: "project", Value: "[[My Project]]", Line: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractInlineFields(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d", len(tt.expected), len(result))
				return
			}
			for i, exp := range tt.expected {
				if result[i].Key != exp.Key {
					t.Errorf("field %d: expected key %q, got %q", i, exp.Key, result[i].Key)
				}
				if result[i].Value != exp.Value {
					t.Errorf("field %d: expected value %q, got %q", i, exp.Value, result[i].Value)
				}
				if result[i].Line != exp.Line {
					t.Errorf("field %d: expected line %d, got %d", i, exp.Line, result[i].Line)
				}
			}
		})
	}
}

func TestSetInlineField(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		key             string
		value           string
		expectedUpdated bool
		shouldContain   string
	}{
		{
			name:            "update existing field",
			content:         "status:: draft\nSome text",
			key:             "status",
			value:           "done",
			expectedUpdated: true,
			shouldContain:   "status:: done",
		},
		{
			name:            "add new field",
			content:         "Some text",
			key:             "status",
			value:           "new",
			expectedUpdated: false,
			shouldContain:   "status:: new",
		},
		{
			name:            "add to note with frontmatter",
			content:         "---\ntitle: Test\n---\n\n# Content",
			key:             "rating",
			value:           "5",
			expectedUpdated: false,
			shouldContain:   "rating:: 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, updated := setInlineField(tt.content, tt.key, tt.value)
			if updated != tt.expectedUpdated {
				t.Errorf("expected updated=%v, got %v", tt.expectedUpdated, updated)
			}
			if tt.shouldContain != "" {
				if !containsString(result, tt.shouldContain) {
					t.Errorf("expected result to contain %q\nresult:\n%s", tt.shouldContain, result)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
