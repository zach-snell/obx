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
	// Matches inline tags: #tag-name
	inlineTagRegex = regexp.MustCompile(`(?:^|[^#\w])#([a-zA-Z0-9_\-]+)`)
	// Matches frontmatter tags array: tags: [tag1, tag2]
	frontmatterArrayRegex = regexp.MustCompile(`tags:\s*\[(.*?)\]`)
	// Matches frontmatter tags list
	frontmatterListRegex = regexp.MustCompile(`tags:\s*\n((?:\s*-\s*.+\n?)+)`)
)

// ExtractTags extracts all tags from note content
func ExtractTags(content string) []string {
	tagSet := make(map[string]bool)

	// Extract frontmatter tags
	if strings.HasPrefix(content, "---") {
		endIdx := strings.Index(content[3:], "---")
		if endIdx > 0 {
			frontmatter := content[3 : endIdx+3]

			// Array format: tags: [tag1, tag2]
			if match := frontmatterArrayRegex.FindStringSubmatch(frontmatter); match != nil {
				for _, tag := range strings.Split(match[1], ",") {
					tag = strings.Trim(strings.TrimSpace(tag), `"'`)
					if tag != "" {
						tagSet[tag] = true
					}
				}
			}

			// List format: tags:\n  - tag1\n  - tag2
			if match := frontmatterListRegex.FindStringSubmatch(frontmatter); match != nil {
				lines := strings.Split(match[1], "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "-") {
						tag := strings.TrimSpace(strings.TrimPrefix(line, "-"))
						if tag != "" {
							tagSet[tag] = true
						}
					}
				}
			}
		}
	}

	// Extract inline tags (skip code blocks)
	contentWithoutCode := regexp.MustCompile("(?s)```.*?```").ReplaceAllString(content, "")
	matches := inlineTagRegex.FindAllStringSubmatch(contentWithoutCode, -1)
	for _, match := range matches {
		tagSet[match[1]] = true
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

// SearchByTagsHandler searches for notes containing all specified tags
func (v *Vault) SearchByTagsHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchTagsArgs) (*mcp.CallToolResult, any, error) {
	tagsStr := args.Tags
	dir := args.Directory

	// Parse comma-separated tags
	var searchTags []string
	for _, tag := range strings.Split(tagsStr, ",") {
		tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if tag != "" {
			searchTags = append(searchTags, strings.ToLower(tag))
		}
	}

	if len(searchTags) == 0 {
		return nil, nil, fmt.Errorf("at least one tag is required")
	}

	searchPath := v.GetPath()
	if dir != "" {
		searchPath = filepath.Join(v.GetPath(), dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	type result struct {
		path string
		tags []string
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

		noteTags := ExtractTags(string(content))

		// Check if note has all search tags (AND operation)
		hasAll := true
		noteTagsLower := make(map[string]bool)
		for _, t := range noteTags {
			noteTagsLower[strings.ToLower(t)] = true
		}
		for _, searchTag := range searchTags {
			if !noteTagsLower[searchTag] {
				hasAll = false
				break
			}
		}

		if hasAll {
			relPath, _ := filepath.Rel(v.GetPath(), path)
			results = append(results, result{path: relPath, tags: noteTags})
		}
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %v", err)
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No notes found with tags: %s", strings.Join(searchTags, ", "))},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes with tags [%s]:\n\n", len(results), strings.Join(searchTags, ", ")))

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("- %s\n", r.path))
		sb.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(r.tags, ", ")))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}
