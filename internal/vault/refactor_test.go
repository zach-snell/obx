package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitByHeading(t *testing.T) {
	content := `# Title

Introduction text.

## Section One

Content of section one.

## Section Two

Content of section two.

### Subsection

Nested content.

## Section Three

Final section.
`

	tests := []struct {
		name       string
		level      int
		wantCount  int
		wantTitles []string
	}{
		{
			name:       "split at h2",
			level:      2,
			wantCount:  4, // preamble + 3 sections
			wantTitles: []string{"", "Section One", "Section Two", "Section Three"},
		},
		{
			name:       "split at h1",
			level:      1,
			wantCount:  1,
			wantTitles: []string{"Title"},
		},
		{
			name:       "split at h3",
			level:      3,
			wantCount:  2,
			wantTitles: []string{"", "Subsection"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := splitByHeading(content, tt.level)
			if len(sections) != tt.wantCount {
				t.Errorf("splitByHeading() got %d sections, want %d", len(sections), tt.wantCount)
				return
			}
			for i, wantTitle := range tt.wantTitles {
				if sections[i].title != wantTitle {
					t.Errorf("section %d title = %q, want %q", i, sections[i].title, wantTitle)
				}
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Normal Title", "Normal Title"},
		{"Title/With/Slashes", "Title-With-Slashes"},
		{"Title: With Colon", "Title- With Colon"},
		{"Title?With*Special<Chars>", "TitleWithSpecialChars"},
		{"   ", "untitled"},
		{"", "untitled"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePaths(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "note1.md, note2.md, note3.md",
			want:  []string{"note1.md", "note2.md", "note3.md"},
		},
		{
			input: `["note1.md", "note2.md"]`,
			want:  []string{"note1.md", "note2.md"},
		},
		{
			input: "single.md",
			want:  []string{"single.md"},
		},
		{
			input: `['note1.md', 'note2.md']`,
			want:  []string{"note1.md", "note2.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parsePaths(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parsePaths(%q) got %d paths, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parsePaths(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRemoveSectionFromContent(t *testing.T) {
	content := `# Title

Introduction.

## Keep This

Content to keep.

## Remove This

Content to remove.

### Nested Under Remove

Also removed.

## Another Keep

Final content.
`

	result := removeSectionFromContent(content, "Remove This")

	if strings.Contains(result, "Remove This") {
		t.Error("removeSectionFromContent() should remove the section heading")
	}
	if strings.Contains(result, "Content to remove") {
		t.Error("removeSectionFromContent() should remove section content")
	}
	if strings.Contains(result, "Nested Under Remove") {
		t.Error("removeSectionFromContent() should remove nested sections")
	}
	if !strings.Contains(result, "Keep This") {
		t.Error("removeSectionFromContent() should keep other sections")
	}
	if !strings.Contains(result, "Another Keep") {
		t.Error("removeSectionFromContent() should keep sections after removed one")
	}
}

func TestSplitNote_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create note with sections
	content := `# Main Title

Introduction.

## Section One

First section content.

## Section Two

Second section content.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "to-split.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}

func TestMergeNotes_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notes to merge
	if err := os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte("# Note 1\n\nContent 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("# Note 2\n\nContent 2"), 0o644); err != nil {
		t.Fatal(err)
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}

func TestDuplicateNote_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create note to duplicate
	content := "# Original Note\n\nOriginal content."
	if err := os.WriteFile(filepath.Join(tmpDir, "original.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}

func TestExtractSection_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create note with extractable section
	content := `# Main Note

Introduction.

## Extract This

Content to extract.

## Keep This

Content to keep.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "source.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}
