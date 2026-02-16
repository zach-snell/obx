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

var (
	// Matches H1 title: # Title
	h1Regex = regexp.MustCompile(`(?m)^#\s+(.+)$`)
)

// MOC represents a Map of Content
type MOC struct {
	Path        string   `json:"path"`
	Title       string   `json:"title"`
	Tags        []string `json:"tags"`
	LinkedNotes []string `json:"linkedNotes"`
}

// ExtractH1Title extracts the first H1 title from content
func ExtractH1Title(content string) string {
	match := h1Regex.FindStringSubmatch(content)
	if match != nil {
		return strings.TrimSpace(match[1])
	}
	return ""
}

// IsMOC checks if a note is a MOC (has #moc tag)
func IsMOC(content string) bool {
	tags := ExtractTags(content)
	for _, tag := range tags {
		if strings.EqualFold(tag, "moc") {
			return true
		}
	}
	return false
}

// DiscoverMOCsHandler discovers all MOCs in the vault
func (v *Vault) DiscoverMOCsHandler(ctx context.Context, req *mcp.CallToolRequest, args DiscoverMOCsArgs) (*mcp.CallToolResult, any, error) {
	dir := args.Directory

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var mocs []MOC

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

		contentStr := string(content)
		if !IsMOC(contentStr) {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		title := ExtractH1Title(contentStr)
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(path), ".md")
		}

		mocs = append(mocs, MOC{
			Path:        relPath,
			Title:       title,
			Tags:        ExtractTags(contentStr),
			LinkedNotes: ExtractWikilinks(contentStr),
		})

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("discovery failed: %v", err)
	}

	if len(mocs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No MOCs found (notes with #moc tag)"},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d MOCs:\n\n", len(mocs)))

	for _, m := range mocs {
		sb.WriteString(fmt.Sprintf("## %s\n", m.Title))
		sb.WriteString(fmt.Sprintf("Path: %s\n", m.Path))
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(m.Tags, ", ")))
		if len(m.LinkedNotes) > 0 {
			sb.WriteString(fmt.Sprintf("Links (%d): %s\n", len(m.LinkedNotes), strings.Join(m.LinkedNotes, ", ")))
		}
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}
