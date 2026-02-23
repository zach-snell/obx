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

// Frontmatter represents parsed YAML frontmatter
type Frontmatter map[string]string

// ParseFrontmatter extracts frontmatter from note content
func ParseFrontmatter(content string) Frontmatter {
	fm := make(Frontmatter)

	if !strings.HasPrefix(content, "---") {
		return fm
	}

	endIdx := strings.Index(content[3:], "---")
	if endIdx < 0 {
		return fm
	}

	fmContent := content[3 : endIdx+3]
	lines := strings.Split(fmContent, "\n")

	// Simple YAML parsing (key: value)
	keyValueRegex := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:\s*(.*)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if match := keyValueRegex.FindStringSubmatch(line); match != nil {
			key := strings.ToLower(match[1])
			value := strings.Trim(strings.TrimSpace(match[2]), `"'`)
			fm[key] = value
		}
	}

	return fm
}

// QueryFrontmatterHandler searches notes by frontmatter properties
func (v *Vault) QueryFrontmatterHandler(ctx context.Context, req *mcp.CallToolRequest, args QueryFrontmatterArgs) (*mcp.CallToolResult, any, error) {
	query := args.Query
	dir := args.Directory

	// Parse query (supports key=value or key:value)
	var key, value string
	if idx := strings.Index(query, "="); idx > 0 {
		key = strings.ToLower(strings.TrimSpace(query[:idx]))
		value = strings.ToLower(strings.TrimSpace(query[idx+1:]))
	} else if idx := strings.Index(query, ":"); idx > 0 {
		key = strings.ToLower(strings.TrimSpace(query[:idx]))
		value = strings.ToLower(strings.TrimSpace(query[idx+1:]))
	} else {
		return nil, nil, fmt.Errorf("invalid query format: use key=value or key:value")
	}

	searchPath := v.GetPath()
	if dir != "" {
		searchPath = filepath.Join(v.GetPath(), dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	type result struct {
		path        string
		frontmatter Frontmatter
	}

	var results []result

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		fm := ParseFrontmatter(string(content))
		if len(fm) == 0 {
			return nil
		}

		// Check if frontmatter matches query
		if fmValue, ok := fm[key]; ok {
			// Support partial matching (contains)
			if strings.Contains(strings.ToLower(fmValue), value) {
				relPath, _ := filepath.Rel(v.GetPath(), path)
				results = append(results, result{path: relPath, frontmatter: fm})
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("query failed: %v", err)
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No notes found matching: %s", query)},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes matching %q:\n\n", len(results), query))

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("## %s\n", r.path))
		for k, v := range r.frontmatter {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// GetFrontmatterHandler returns frontmatter for a specific note
func (v *Vault) GetFrontmatterHandler(ctx context.Context, req *mcp.CallToolRequest, args GetFrontmatterArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path

	if !strings.HasSuffix(path, ".md") {
		return nil, nil, fmt.Errorf("path must end with .md")
	}

	fullPath := filepath.Join(v.GetPath(), path)

	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("note not found: %s", path)
		}
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}

	fm := ParseFrontmatter(string(content))

	if len(fm) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No frontmatter found in: %s", path)},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Frontmatter for %s:\n\n", path))
	for k, v := range fm {
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}
