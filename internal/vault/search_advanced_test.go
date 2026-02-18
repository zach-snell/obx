package vault

import (
	"testing"
)

func TestParseSearchTerms(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "simple terms",
			query: "simple terms",
			want:  []string{"simple", "terms"},
		},
		{
			name:  "quoted phrase",
			query: `"quoted phrase" other`,
			want:  []string{"quoted phrase", "other"},
		},
		{
			name:  "terms with operators",
			query: `term1 AND term2 OR term3`,
			want:  []string{"term1", "AND", "term2", "OR", "term3"},
		},
		{
			name:  "multiple quoted phrases",
			query: `"first phrase" "second phrase"`,
			want:  []string{"first phrase", "second phrase"},
		},
		{
			name:  "single character terms skipped",
			query: "a", // single character should be skipped by filter
			want:  []string{},
		},
		{
			name:  "test NOT excluded",
			query: "test NOT excluded",
			want:  []string{"test", "NOT", "excluded"},
		},
		{
			name:  "quoted multi word",
			query: `"multi word phrase"`,
			want:  []string{"multi word phrase"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSearchTerms(tt.query)
			if len(got) != len(tt.want) {
				t.Errorf("parseSearchTerms(%q) = %v, want %v", tt.query, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseSearchTerms(%q)[%d] = %q, want %q", tt.query, i, got[i], tt.want[i])
				}
			}
		})
	}
}
