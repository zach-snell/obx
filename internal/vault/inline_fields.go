package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InlineField represents a Dataview-style inline field
type InlineField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Line  int    `json:"line"`
}

// inlineFieldRegex matches Dataview inline fields:
// - Standard: key:: value
// - With brackets: [key:: value] (key visible)
// - With parens: (key:: value) (key hidden in reading mode)
var (
	// key:: value (entire line or end of line)
	standardFieldRegex = regexp.MustCompile(`(?:^|\s)([a-zA-Z_][a-zA-Z0-9_-]*)\s*::\s*(.+?)(?:$|\n)`)
	// [key:: value]
	bracketFieldRegex = regexp.MustCompile(`\[([a-zA-Z_][a-zA-Z0-9_-]*)\s*::\s*([^\]]+)\]`)
	// (key:: value)
	parenFieldRegex = regexp.MustCompile(`\(([a-zA-Z_][a-zA-Z0-9_-]*)\s*::\s*([^)]+)\)`)
)

// ExtractInlineFields extracts all inline fields from content
func ExtractInlineFields(content string) []InlineField {
	var fields []InlineField
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		// Try all patterns
		for _, match := range standardFieldRegex.FindAllStringSubmatch(line, -1) {
			fields = append(fields, InlineField{
				Key:   strings.TrimSpace(match[1]),
				Value: strings.TrimSpace(match[2]),
				Line:  lineNum + 1,
			})
		}
		for _, match := range bracketFieldRegex.FindAllStringSubmatch(line, -1) {
			fields = append(fields, InlineField{
				Key:   strings.TrimSpace(match[1]),
				Value: strings.TrimSpace(match[2]),
				Line:  lineNum + 1,
			})
		}
		for _, match := range parenFieldRegex.FindAllStringSubmatch(line, -1) {
			fields = append(fields, InlineField{
				Key:   strings.TrimSpace(match[1]),
				Value: strings.TrimSpace(match[2]),
				Line:  lineNum + 1,
			})
		}
	}

	return fields
}

// GetInlineFieldsHandler extracts inline fields from a note
func (v *Vault) GetInlineFieldsHandler(ctx context.Context, req *mcp.CallToolRequest, args GetInlineFieldsArgs) (*mcp.CallToolResult, any, error) {
	notePath := args.Path

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

	fields := ExtractInlineFields(string(content))

	if len(fields) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No inline fields found in: %s", notePath)},
			},
		}, nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Inline Fields in %s\n\n", notePath)
	fmt.Fprintf(&sb, "Found %d fields:\n\n", len(fields))

	for _, f := range fields {
		fmt.Fprintf(&sb, "- **%s**: %s (L%d)\n", f.Key, f.Value, f.Line)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// SetInlineFieldHandler sets or updates an inline field in a note
func (v *Vault) SetInlineFieldHandler(ctx context.Context, req *mcp.CallToolRequest, args SetInlineFieldArgs) (*mcp.CallToolResult, any, error) {
	notePath := args.Path
	key := args.Key
	value := args.Value

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
	newContent, updated := setInlineField(contentStr, key, value)

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	action := "Set"
	if updated {
		action = "Updated"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%s %s:: %s in %s", action, key, value, notePath)},
		},
	}, nil, nil
}

// inlineFieldQuery holds query parameters
type inlineFieldQuery struct {
	key      string
	value    string
	operator string
}

// matchesQuery checks if a field matches the query
func (q *inlineFieldQuery) matches(f InlineField) bool {
	if !strings.EqualFold(f.Key, q.key) {
		return false
	}
	switch q.operator {
	case "exists":
		return true
	case "equals":
		return strings.EqualFold(f.Value, q.value)
	default: // contains
		return q.value == "" || strings.Contains(strings.ToLower(f.Value), strings.ToLower(q.value))
	}
}

// description returns a human-readable query description
func (q *inlineFieldQuery) description() string {
	if q.value == "" {
		return q.key
	}
	return fmt.Sprintf("%s %s '%s'", q.key, q.operator, q.value)
}

// inlineFieldResult holds a query result
type inlineFieldResult struct {
	path   string
	fields []InlineField
}

// searchInlineFields searches the vault for inline fields matching the query
func (v *Vault) searchInlineFields(dir string, query *inlineFieldQuery) ([]inlineFieldResult, error) {
	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, fmt.Errorf("search path must be within vault")
	}

	var results []inlineFieldResult

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		fields := ExtractInlineFields(string(content))
		var matched []InlineField

		for _, f := range fields {
			if query.matches(f) {
				matched = append(matched, f)
			}
		}

		if len(matched) > 0 {
			relPath, _ := filepath.Rel(v.path, path)
			results = append(results, inlineFieldResult{path: relPath, fields: matched})
		}

		return nil
	})

	return results, err
}

// QueryInlineFieldsHandler searches notes by inline field values
func (v *Vault) QueryInlineFieldsHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchInlineFieldsArgs) (*mcp.CallToolResult, any, error) {
	key := args.Key
	value := args.Value
	operator := args.Operator
	dir := args.Directory

	if operator == "" {
		operator = "contains"
	}

	query := &inlineFieldQuery{
		key:      key,
		value:    value,
		operator: operator,
	}

	results, err := v.searchInlineFields(dir, query)
	if err != nil {
		return nil, nil, fmt.Errorf("query failed: %v", err)
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No notes found with inline field: %s", query.description())},
			},
		}, nil, nil
	}

	var sb strings.Builder
	queryDesc := query.description()
	fmt.Fprintf(&sb, "# Notes with %s\n\n", queryDesc)
	fmt.Fprintf(&sb, "Found %d notes:\n\n", len(results))

	for _, r := range results {
		fmt.Fprintf(&sb, "## %s\n", r.path)
		for _, f := range r.fields {
			fmt.Fprintf(&sb, "- L%d: %s:: %s\n", f.Line, f.Key, f.Value)
		}
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// setInlineField updates or appends an inline field
func setInlineField(content, key, value string) (string, bool) {
	lines := strings.Split(content, "\n")
	keyLower := strings.ToLower(key)

	// Try to find and update existing field
	for i, line := range lines {
		// Check all patterns for existing field with this key
		patterns := []*regexp.Regexp{
			regexp.MustCompile(`(?:^|\s)(` + regexp.QuoteMeta(key) + `)\s*::\s*.+`),
			regexp.MustCompile(`\[(` + regexp.QuoteMeta(key) + `)\s*::\s*[^\]]+\]`),
			regexp.MustCompile(`\((` + regexp.QuoteMeta(key) + `)\s*::\s*[^)]+\)`),
		}

		for _, pattern := range patterns {
			if match := pattern.FindStringSubmatch(line); match != nil {
				if strings.EqualFold(match[1], keyLower) {
					// Replace the value
					newField := fmt.Sprintf("%s:: %s", key, value)
					lines[i] = pattern.ReplaceAllString(line, " "+newField)
					lines[i] = strings.TrimPrefix(lines[i], " ")
					return strings.Join(lines, "\n"), true
				}
			}
		}
	}

	// Field not found, append at end
	// Find a good place to insert - after frontmatter or at end
	if strings.HasPrefix(content, "---") {
		// Has frontmatter, insert after it
		fmEnd := strings.Index(content[3:], "---")
		if fmEnd > 0 {
			fmEnd += 6 // Account for first --- and second ---
			// Skip any blank lines after frontmatter
			for fmEnd < len(content) && (content[fmEnd] == '\n' || content[fmEnd] == '\r') {
				fmEnd++
			}
			newContent := content[:fmEnd] + fmt.Sprintf("\n%s:: %s\n", key, value) + content[fmEnd:]
			return newContent, false
		}
	}

	// No frontmatter or couldn't parse, append at end
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + fmt.Sprintf("%s:: %s\n", key, value), false
}
