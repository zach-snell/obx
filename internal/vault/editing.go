package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// maxContextLineChars is the maximum characters per line in context output
	maxContextLineChars = 200
	// maxContextContentPreview is the max lines to show from each edge of inserted/replaced content
	maxContextContentPreview = 2
)

var (
	// headingRegex matches markdown headings: ## Heading
	headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
)

// truncateLine truncates a line to maxChars, adding an indicator if truncated
func truncateLine(line string, maxChars int) string {
	if len(line) <= maxChars {
		return line
	}
	return line[:maxChars] + fmt.Sprintf("... [%d chars total]", len(line))
}

// buildEditContext builds a context snippet around an edit location.
// editStart and editEnd are 0-based line indices of the edited region.
// insertedLines are the lines that were inserted/replaced (may be nil).
// allLines are the full file lines AFTER the edit.
// contextN is the number of surrounding lines to show.
func buildEditContext(allLines []string, editStart, editEnd, contextN int, insertedLines []string) string {
	if contextN <= 0 {
		return ""
	}

	var sb strings.Builder

	// Lines before the edit
	beforeStart := editStart - contextN
	if beforeStart < 0 {
		beforeStart = 0
	}
	for i := beforeStart; i < editStart; i++ {
		fmt.Fprintf(&sb, "L%d: %s\n", i+1, truncateLine(allLines[i], maxContextLineChars))
	}

	// The edit region itself
	editLen := editEnd - editStart
	switch {
	case editLen <= 0 && len(insertedLines) > 0:
		// Insertion point (no lines replaced, just inserted)
		writeContentPreview(&sb, insertedLines, editStart, "INSERTED")
	case len(insertedLines) > 0:
		// Replacement
		writeContentPreview(&sb, insertedLines, editStart, "CHANGED")
	default:
		// Show the actual edit lines
		for i := editStart; i < editEnd && i < len(allLines); i++ {
			fmt.Fprintf(&sb, "L%d: %s\n", i+1, truncateLine(allLines[i], maxContextLineChars))
		}
	}

	// Lines after the edit
	afterEnd := editEnd + contextN
	if afterEnd > len(allLines) {
		afterEnd = len(allLines)
	}
	for i := editEnd; i < afterEnd; i++ {
		fmt.Fprintf(&sb, "L%d: %s\n", i+1, truncateLine(allLines[i], maxContextLineChars))
	}

	return sb.String()
}

// writeContentPreview writes a compact preview of inserted/changed content
func writeContentPreview(sb *strings.Builder, contentLines []string, startLine int, label string) {
	if len(contentLines) <= maxContextContentPreview*2+1 {
		// Short enough to show fully
		for i, line := range contentLines {
			marker := ""
			if i == 0 {
				marker = fmt.Sprintf("  ← %s", label)
			}
			fmt.Fprintf(sb, "L%d: %s%s\n", startLine+i+1, truncateLine(line, maxContextLineChars), marker)
		}
	} else {
		// Show first N and last N lines
		for i := 0; i < maxContextContentPreview; i++ {
			marker := ""
			if i == 0 {
				marker = fmt.Sprintf("  ← %s", label)
			}
			fmt.Fprintf(sb, "L%d: %s%s\n", startLine+i+1, truncateLine(contentLines[i], maxContextLineChars), marker)
		}
		fmt.Fprintf(sb, "     [... %d more lines ...]\n", len(contentLines)-maxContextContentPreview*2)
		for i := len(contentLines) - maxContextContentPreview; i < len(contentLines); i++ {
			fmt.Fprintf(sb, "L%d: %s\n", startLine+i+1, truncateLine(contentLines[i], maxContextLineChars))
		}
	}
}

// EditNoteHandler performs surgical find-and-replace within a note
func (v *Vault) EditNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args EditNoteArgs) (*mcp.CallToolResult, any, error) {
	notePath := args.Path
	oldText := args.OldText
	replacementText := args.NewText
	replaceAll := args.ReplaceAll
	contextN := args.ContextLines

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.path, notePath)
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("note not found: %s", notePath)
		}
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}

	contentStr := string(content)

	// Count occurrences
	count := strings.Count(contentStr, oldText)
	if count == 0 {
		return nil, nil, fmt.Errorf("old_text not found in %s", notePath)
	}

	if count > 1 && !replaceAll {
		return nil, nil, fmt.Errorf(
			"found %d occurrences of old_text in %s. Use replace_all=true to replace all, or provide more context to match uniquely",
			count, notePath,
		)
	}

	// Perform replacement
	var newContent string
	replaced := 0
	if replaceAll {
		newContent = strings.ReplaceAll(contentStr, oldText, replacementText)
		replaced = count
	} else {
		newContent = strings.Replace(contentStr, oldText, replacementText, 1)
		replaced = 1
	}

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	// Build response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Replaced %d occurrence(s) in %s", replaced, notePath)

	if contextN > 0 {
		// Find the first edit location for context
		newLines := strings.Split(newContent, "\n")
		replacementLines := strings.Split(replacementText, "\n")

		// Find where the replacement starts
		idx := strings.Index(newContent, replacementText)
		if idx >= 0 {
			editStartLine := strings.Count(newContent[:idx], "\n")
			editEndLine := editStartLine + len(replacementLines)

			ctxStr := buildEditContext(newLines, editStartLine, editEndLine, contextN, replacementLines)
			fmt.Fprintf(&sb, "\n\n--- Context ---\n%s", ctxStr)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// ReplaceSectionHandler replaces content under a heading
func (v *Vault) ReplaceSectionHandler(ctx context.Context, req *mcp.CallToolRequest, args ReplaceSectionArgs) (*mcp.CallToolResult, any, error) {
	notePath := args.Path
	heading := args.Heading
	sectionContent := args.Content
	contextN := args.ContextLines

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.path, notePath)
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("note not found: %s", notePath)
		}
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the section boundaries
	sectionStart := -1 // Line index of heading
	sectionEnd := -1   // Line index of next heading (exclusive)
	sectionLevel := 0

	for i, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			level := len(matches[1])
			text := strings.TrimSpace(matches[2])

			if sectionStart == -1 && strings.EqualFold(text, heading) {
				sectionStart = i
				sectionLevel = level
				continue
			}

			if sectionStart >= 0 && level <= sectionLevel {
				sectionEnd = i
				break
			}
		}
	}

	if sectionStart == -1 {
		return nil, nil, fmt.Errorf("heading '%s' not found in %s", heading, notePath)
	}

	if sectionEnd == -1 {
		sectionEnd = len(lines)
	}

	// Content starts after the heading line
	contentStart := sectionStart + 1
	linesReplaced := sectionEnd - contentStart

	// Normalize new content: ensure blank line after heading, trim trailing newlines
	normalizedContent := strings.TrimRight(sectionContent, "\n")
	newContentLines := strings.Split("\n"+normalizedContent+"\n", "\n")

	// Build new file: before heading + heading + new content + after section
	var result []string
	result = append(result, lines[:contentStart]...) // Up to and including heading
	result = append(result, newContentLines...)      // New section content
	result = append(result, lines[sectionEnd:]...)   // Everything after section

	finalContent := strings.Join(result, "\n")

	if err := os.WriteFile(fullPath, []byte(finalContent), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	// Build response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Replaced section '%s' in %s (%d lines replaced with %d lines)",
		heading, notePath, linesReplaced, len(newContentLines))

	if contextN > 0 {
		finalLines := strings.Split(finalContent, "\n")
		editEnd := contentStart + len(newContentLines)
		if editEnd > len(finalLines) {
			editEnd = len(finalLines)
		}
		ctxStr := buildEditContext(finalLines, contentStart, editEnd, contextN, newContentLines)
		fmt.Fprintf(&sb, "\n\n--- Context ---\n%s", ctxStr)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// AppendNoteHandler appends content to a note, optionally at a specific position
func (v *Vault) AppendNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args AppendNoteArgs) (*mcp.CallToolResult, any, error) {
	notePath := args.Path
	content := args.Content
	position := args.Position
	after := args.After
	before := args.Before
	contextN := args.ContextLines

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.path, notePath)
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	// Create if not exists (only for default append or simple path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return nil, nil, fmt.Errorf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			return nil, nil, fmt.Errorf("failed to write note: %v", err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Created new note: %s", notePath)},
			},
		}, nil, nil
	}

	fileContent, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}
	lines := strings.Split(string(fileContent), "\n")
	newLines := strings.Split(content, "\n")

	var insertIndex int
	var insertMode string // "insert" or "append"

	switch position {
	case "start":
		insertIndex = 0
		insertMode = "insert"
	case "after":
		if after == "" {
			return nil, nil, fmt.Errorf("argument 'after' is required when position is 'after'")
		}
		idx, err := findTargetLine(lines, after)
		if err != nil {
			return nil, nil, err
		}
		insertIndex = idx + 1
		insertMode = "insert"
	case "before":
		if before == "" {
			return nil, nil, fmt.Errorf("argument 'before' is required when position is 'before'")
		}
		idx, err := findTargetLine(lines, before)
		if err != nil {
			return nil, nil, err
		}
		insertIndex = idx
		insertMode = "insert"
	default: // "end" or empty
		insertIndex = len(lines)
		insertMode = "append"
	}

	// Construct new content
	finalLines := buildFinalLines(lines, newLines, insertIndex, insertMode)

	finalContent := strings.Join(finalLines, "\n")
	if err := os.WriteFile(fullPath, []byte(finalContent), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	// Build response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Appended content to %s", notePath)

	if contextN > 0 {
		editStart := insertIndex
		if insertMode == "append" {
			// For append, context is end of original + new
			editStart = len(lines)
			if len(lines) > 0 && lines[len(lines)-1] != "" {
				editStart++ // account for added newline
			}
		}
		editEnd := editStart + len(newLines)
		ctxStr := buildEditContext(finalLines, editStart, editEnd, contextN, newLines)
		fmt.Fprintf(&sb, "\n\n--- Context ---\n%s", ctxStr)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// buildFinalLines constructs the final line slice for an append/insert operation.
func buildFinalLines(lines, newLines []string, insertIndex int, insertMode string) []string {
	if insertMode == "append" {
		var base []string
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			base = make([]string, len(lines), len(lines)+1+len(newLines))
			copy(base, lines)
			base = append(base, "")
		} else {
			base = make([]string, len(lines), len(lines)+len(newLines))
			copy(base, lines)
		}
		return append(base, newLines...)
	}

	// Insert at index
	if insertIndex > len(lines) {
		insertIndex = len(lines)
	}
	result := make([]string, 0, len(lines)+len(newLines))
	result = append(result, lines[:insertIndex]...)
	result = append(result, newLines...)
	result = append(result, lines[insertIndex:]...)
	return result
}

// findTargetLine finds the line index matching the target string (heading or text)
func findTargetLine(lines []string, target string) (int, error) {
	targetLower := strings.ToLower(target)
	var matches []int

	for i, line := range lines {
		// Check for heading match first
		if headingMatch := headingRegex.FindStringSubmatch(line); headingMatch != nil {
			if strings.EqualFold(strings.TrimSpace(headingMatch[2]), target) {
				return i, nil // Exact heading match is prioritized
			}
		}

		// Check for content match
		if strings.Contains(strings.ToLower(line), targetLower) {
			matches = append(matches, i)
		}
	}

	if len(matches) == 0 {
		return -1, fmt.Errorf("target '%s' not found", target)
	}
	if len(matches) > 1 {
		return -1, fmt.Errorf("ambiguous target '%s' found %d times", target, len(matches))
	}

	return matches[0], nil
}

// locatedEdit is an editEntry with its byte offset in the file content
type locatedEdit struct {
	EditEntry
	offset int // byte offset of old_text in content
}

// BatchEditNoteHandler performs multiple find-and-replace operations atomically
func (v *Vault) BatchEditNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args BatchEditArgs) (*mcp.CallToolResult, any, error) {
	notePath := args.Path
	edits := args.Edits
	contextN := args.ContextLines

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.path, notePath)
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	if len(edits) == 0 {
		return nil, nil, fmt.Errorf("edits array is empty")
	}

	// Read file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("note not found: %s", notePath)
		}
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}

	contentStr := string(content)

	// Validate all edits before applying any
	located, validationErr := validateBatchEdits(contentStr, edits, notePath)
	if validationErr != nil {
		return nil, nil, validationErr
	}

	// Sort by offset descending so we can apply from end to start without shifting
	sort.Slice(located, func(i, j int) bool {
		return located[i].offset > located[j].offset
	})

	// Apply all edits
	result := contentStr
	for _, le := range located {
		result = result[:le.offset] + le.NewText + result[le.offset+len(le.OldText):]
	}

	if err := os.WriteFile(fullPath, []byte(result), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	// Build response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Applied %d edit(s) to %s", len(edits), notePath)

	if contextN > 0 && len(located) > 0 {
		// Show context around the first edit (by file position, which is last in our sorted slice)
		firstEdit := located[len(located)-1]
		newLines := strings.Split(result, "\n")
		replacementLines := strings.Split(firstEdit.NewText, "\n")

		idx := strings.Index(result, firstEdit.NewText)
		if idx >= 0 {
			editStartLine := strings.Count(result[:idx], "\n")
			editEndLine := editStartLine + len(replacementLines)
			ctxStr := buildEditContext(newLines, editStartLine, editEndLine, contextN, replacementLines)
			fmt.Fprintf(&sb, "\n\n--- Context (first edit) ---\n%s", ctxStr)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// validateBatchEdits checks that every old_text exists exactly once and that no edits overlap.
// Returns located edits with byte offsets, or an error result.
func validateBatchEdits(content string, edits []EditEntry, notePath string) ([]locatedEdit, error) {
	var located []locatedEdit
	var errors []string

	for i, e := range edits {
		if e.OldText == "" {
			errors = append(errors, fmt.Sprintf("edit %d: old_text is empty", i+1))
			continue
		}

		count := strings.Count(content, e.OldText)
		switch {
		case count == 0:
			errors = append(errors, fmt.Sprintf("edit %d: old_text not found: %q", i+1, truncateLine(e.OldText, 80)))
		case count > 1:
			errors = append(errors, fmt.Sprintf("edit %d: old_text found %d times (must be unique): %q", i+1, count, truncateLine(e.OldText, 80)))
		default:
			offset := strings.Index(content, e.OldText)
			located = append(located, locatedEdit{EditEntry: e, offset: offset})
		}
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf(
			"batch edit validation failed for %s:\n- %s",
			notePath, strings.Join(errors, "\n- "),
		)
	}

	// Check for overlapping edits
	sort.Slice(located, func(i, j int) bool {
		return located[i].offset < located[j].offset
	})
	for i := 1; i < len(located); i++ {
		prevEnd := located[i-1].offset + len(located[i-1].OldText)
		if located[i].offset < prevEnd {
			return nil, fmt.Errorf(
				"batch edit validation failed for %s: edits %d and %d overlap",
				notePath, i, i+1,
			)
		}
	}

	return located, nil
}
