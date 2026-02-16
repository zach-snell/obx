package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchResult represents a search match
type SearchResult struct {
	File    string
	Line    int
	Content string
}

// SearchVaultHandler searches for content in vault notes
func (v *Vault) SearchVaultHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
	query := args.Query
	dir := args.Directory

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	queryLower := strings.ToLower(query)
	var results []SearchResult

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

		lines := strings.Split(string(content), "\n")
		relPath, _ := filepath.Rel(v.path, path)

		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), queryLower) {
				results = append(results, SearchResult{
					File:    relPath,
					Line:    i + 1,
					Content: strings.TrimSpace(line),
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %v", err)
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No matches found for: %s", query)},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d matches for %q:\n\n", len(results), query))

	currentFile := ""
	for _, r := range results {
		if r.File != currentFile {
			if currentFile != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("## %s\n", r.File))
			currentFile = r.File
		}
		sb.WriteString(fmt.Sprintf("  L%d: %s\n", r.Line, truncate(r.Content, 100)))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
