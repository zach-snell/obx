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
	// Matches wikilinks: [[Note Name]] or [[path/to/note|Alias]]
	wikilinkRegex = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
)

// ExtractWikilinks extracts all wikilinks from content
func ExtractWikilinks(content string) []string {
	matches := wikilinkRegex.FindAllStringSubmatch(content, -1)
	var links []string
	seen := make(map[string]bool)
	for _, match := range matches {
		link := strings.TrimSpace(match[1])
		if link != "" && !seen[link] {
			links = append(links, link)
			seen[link] = true
		}
	}
	return links
}

// BacklinksHandler finds all notes that link to a given note
func (v *Vault) BacklinksHandler(ctx context.Context, req *mcp.CallToolRequest, args GetBacklinksArgs) (*mcp.CallToolResult, any, error) {
	target := args.Path

	// Normalize target: remove .md extension for matching
	targetName := strings.TrimSuffix(target, ".md")
	targetBase := filepath.Base(targetName)

	// Build regex patterns for wikilinks
	// Matches: [[target]], [[target|alias]], [[path/to/target]], [[path/to/target|alias]]
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\[\[` + regexp.QuoteMeta(targetName) + `(\|[^\]]+)?\]\]`),
		regexp.MustCompile(`\[\[` + regexp.QuoteMeta(targetBase) + `(\|[^\]]+)?\]\]`),
	}

	type backlink struct {
		path    string
		count   int
		context []string
	}

	var backlinks []backlink

	err := filepath.Walk(v.GetPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(v.GetPath(), path)
		// Skip the target note itself
		if relPath == target || strings.TrimSuffix(relPath, ".md") == targetName {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		var matches []string
		matchCount := 0

		for _, pattern := range patterns {
			for i, line := range lines {
				if pattern.MatchString(line) {
					matchCount++
					// Add context (truncated line)
					ctxLine := strings.TrimSpace(line)
					if len(ctxLine) > 100 {
						ctxLine = ctxLine[:100] + "..."
					}
					matches = append(matches, fmt.Sprintf("L%d: %s", i+1, ctxLine))
				}
			}
		}

		if matchCount > 0 {
			backlinks = append(backlinks, backlink{
				path:    relPath,
				count:   matchCount,
				context: matches,
			})
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to search backlinks: %v", err)
	}

	if len(backlinks) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No backlinks found for: %s", target)},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes linking to %s:\n\n", len(backlinks), target))

	for _, bl := range backlinks {
		sb.WriteString(fmt.Sprintf("## %s (%d links)\n", bl.path, bl.count))
		for _, ctxLine := range bl.context {
			sb.WriteString(fmt.Sprintf("  %s\n", ctxLine))
		}
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// updateLinksInVault updates wikilinks from oldName to newName across the vault
func (v *Vault) updateLinksInVault(oldName, newName string) error {
	// Prepare link patterns
	// Handle [[oldName]] and [[oldName|alias]]
	patterns := []struct {
		pattern *regexp.Regexp
		replace string
	}{
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldName) + `\]\]`),
			replace: "[[" + newName + "]]",
		},
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldName) + `\|([^\]]+)\]\]`),
			replace: "[[" + newName + "|$1]]",
		},
	}

	return filepath.Walk(v.GetPath(), func(path string, info os.FileInfo, err error) error {
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
		newContent := contentStr

		for _, r := range patterns {
			newContent = r.pattern.ReplaceAllString(newContent, r.replace)
		}

		if newContent != contentStr {
			if err := os.WriteFile(path, []byte(newContent), 0o600); err != nil {
				return nil // Continue on error
			}
		}

		return nil
	})
}

// RenameNoteHandler renames a note and updates all links to it
func (v *Vault) RenameNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args RenameNoteArgs) (*mcp.CallToolResult, any, error) {
	oldPath := args.OldPath
	newPath := args.NewPath

	if !strings.HasSuffix(oldPath, ".md") || !strings.HasSuffix(newPath, ".md") {
		return nil, nil, fmt.Errorf("paths must end with .md")
	}

	oldFullPath := filepath.Join(v.GetPath(), oldPath)
	newFullPath := filepath.Join(v.GetPath(), newPath)

	if !v.isPathSafe(oldFullPath) || !v.isPathSafe(newFullPath) {
		return nil, nil, fmt.Errorf("paths must be within vault")
	}

	// Check source exists
	if _, err := os.Stat(oldFullPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("note not found: %s", oldPath)
	}

	// Check destination doesn't exist
	if _, err := os.Stat(newFullPath); err == nil {
		return nil, nil, fmt.Errorf("destination already exists: %s", newPath)
	}

	// Prepare link patterns
	oldName := strings.TrimSuffix(oldPath, ".md")
	oldBase := strings.TrimSuffix(filepath.Base(oldPath), ".md")
	newName := strings.TrimSuffix(newPath, ".md")
	newBase := strings.TrimSuffix(filepath.Base(newPath), ".md")

	// Patterns to find and replace
	replacements := []struct {
		pattern *regexp.Regexp
		replace string
	}{
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldName) + `\]\]`),
			replace: "[[" + newName + "]]",
		},
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldName) + `\|([^\]]+)\]\]`),
			replace: "[[" + newName + "|$1]]",
		},
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldBase) + `\]\]`),
			replace: "[[" + newBase + "]]",
		},
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldBase) + `\|([^\]]+)\]\]`),
			replace: "[[" + newBase + "|$1]]",
		},
	}

	// Update all notes that link to the old path
	updatedFiles := 0
	err := filepath.Walk(v.GetPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if path == oldFullPath {
			return nil // Skip the file we're renaming
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		newContent := contentStr

		for _, r := range replacements {
			newContent = r.pattern.ReplaceAllString(newContent, r.replace)
		}

		if newContent != contentStr {
			if err := os.WriteFile(path, []byte(newContent), 0o600); err != nil {
				return nil // Continue on error
			}
			updatedFiles++
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to update links: %v", err)
	}

	// Create destination directory if needed
	newDir := filepath.Dir(newFullPath)
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Rename the file
	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		return nil, nil, fmt.Errorf("failed to rename note: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Renamed %s -> %s\nUpdated links in %d files", oldPath, newPath, updatedFiles)},
		},
	}, nil, nil
}

// updateWikilinks replaces wikilinks from oldName to newName
func updateWikilinks(content, oldName, newName string) string {
	// Handle [[oldName]] and [[oldName|alias]]
	patterns := []struct {
		old string
		new string
	}{
		{"[[" + oldName + "]]", "[[" + newName + "]]"},
		{"[[" + oldName + "|", "[[" + newName + "|"},
	}

	for _, p := range patterns {
		content = strings.ReplaceAll(content, p.old, p.new)
	}

	return content
}
