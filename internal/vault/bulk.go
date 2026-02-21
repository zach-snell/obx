package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// hasFrontmatter checks if content has YAML frontmatter
func hasFrontmatter(content string) bool {
	return strings.HasPrefix(content, "---")
}

// BulkTagHandler adds or removes tags from multiple notes
func (v *Vault) BulkTagHandler(ctx context.Context, req *mcp.CallToolRequest, args BulkTagArgs) (*mcp.CallToolResult, any, error) {
	pathsStr := args.Paths
	tag := args.Tag
	action := args.Action // add, remove
	dryRun := args.DryRun

	if action == "" {
		action = "add"
	}

	// Normalize tag (remove # prefix if present)
	tag = strings.TrimPrefix(tag, "#")

	paths := parsePaths(pathsStr)
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("at least one path is required")
	}

	var results []string
	var errors []string

	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			p += ".md"
		}

		fullPath := filepath.Join(v.path, p)
		if !v.isPathSafe(fullPath) {
			errors = append(errors, fmt.Sprintf("%s: path must be within vault", p))
			continue
		}
		content, err := os.ReadFile(fullPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read failed", p))
			continue
		}

		contentStr := string(content)
		var modified bool

		if action == "remove" {
			modified, contentStr = removeTagFromNote(contentStr, tag)
		} else {
			modified, contentStr = addTagToNote(contentStr, tag)
		}

		if !modified {
			results = append(results, fmt.Sprintf("%s: no change", p))
			continue
		}

		if !dryRun {
			if err := os.WriteFile(fullPath, []byte(contentStr), 0o600); err != nil {
				errors = append(errors, fmt.Sprintf("%s: write failed", p))
				continue
			}
		}

		if dryRun {
			results = append(results, fmt.Sprintf("%s: would be %sed #%s", p, action, tag))
		} else {
			results = append(results, fmt.Sprintf("%s: %sed #%s", p, action, tag))
		}
	}

	var sb strings.Builder
	if action == "add" {
		if dryRun {
			sb.WriteString(fmt.Sprintf("# Dry Run: Bulk Add Tag: #%s\n\n", tag))
		} else {
			sb.WriteString(fmt.Sprintf("# Bulk Add Tag: #%s\n\n", tag))
		}
	} else {
		if dryRun {
			sb.WriteString(fmt.Sprintf("# Dry Run: Bulk Remove Tag: #%s\n\n", tag))
		} else {
			sb.WriteString(fmt.Sprintf("# Bulk Remove Tag: #%s\n\n", tag))
		}
	}

	if len(results) > 0 {
		sb.WriteString("## Results\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	if len(errors) > 0 {
		sb.WriteString("\n## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// addTagToNote adds a tag to a note's content
func addTagToNote(content, tag string) (modified bool, result string) {
	// Check if tag already exists
	existingTags := ExtractTags(content)
	for _, t := range existingTags {
		if strings.EqualFold(t, tag) {
			return false, content
		}
	}

	// Try to add to frontmatter first
	if hasFrontmatter(content) {
		fm := ParseFrontmatter(content)
		if tagsVal, ok := fm["tags"]; ok {
			// Append to existing tags
			newTags := tagsVal + ", " + tag
			return true, setFrontmatterKey(content, "tags", newTags)
		}
		// Add tags field
		return true, addFrontmatterField(content, "tags", tag)
	}

	// Add inline tag at the end
	return true, content + "\n\n#" + tag
}

// removeTagFromNote removes a tag from a note
func removeTagFromNote(content, tag string) (changed bool, newContent string) {
	changed = false
	lines := strings.Split(content, "\n")
	var resultLines []string

	for _, line := range lines {
		newLine := line
		patterns := []string{
			"#" + tag + " ",
			" #" + tag + " ",
			"#" + tag + "\n",
			" #" + tag + "\n",
			"#" + tag,
		}

		for _, pattern := range patterns {
			if strings.Contains(newLine, pattern) {
				newLine = strings.ReplaceAll(newLine, pattern, " ")
				changed = true
			}
		}

		if strings.HasSuffix(strings.TrimSpace(newLine), "#"+tag) {
			idx := strings.LastIndex(newLine, "#"+tag)
			newLine = strings.TrimSpace(newLine[:idx])
			changed = true
		}

		resultLines = append(resultLines, newLine)
	}

	newContent = strings.Join(resultLines, "\n")

	if hasFrontmatter(newContent) {
		fm := ParseFrontmatter(newContent)
		if tagsVal, ok := fm["tags"]; ok {
			tagList := strings.Split(tagsVal, ",")
			var newTags []string
			for _, t := range tagList {
				t = strings.TrimSpace(t)
				t = strings.Trim(t, "[]")
				if !strings.EqualFold(t, tag) {
					newTags = append(newTags, t)
				} else {
					changed = true
				}
			}
			if len(newTags) > 0 {
				newContent = setFrontmatterKey(newContent, "tags", strings.Join(newTags, ", "))
			} else {
				newContent, _ = removeFrontmatterKey(newContent, "tags")
			}
		}
	}

	return changed, newContent
}

// addFrontmatterField adds a new field to existing frontmatter
func addFrontmatterField(content, key, value string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inFrontmatter := false
	fieldAdded := false

	for i, line := range lines {
		if i == 0 && line == "---" {
			inFrontmatter = true
			result = append(result, line)
			continue
		}

		if inFrontmatter && line == "---" {
			// Add field before closing ---
			if !fieldAdded {
				result = append(result, fmt.Sprintf("%s: %s", key, value))
				fieldAdded = true
			}
			inFrontmatter = false
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// BulkMoveHandler moves multiple notes to a folder
func (v *Vault) BulkMoveHandler(ctx context.Context, req *mcp.CallToolRequest, args BulkMoveArgs) (*mcp.CallToolResult, any, error) {
	pathsStr := args.Paths
	destination := args.Destination
	updateLinks := args.UpdateLinks
	dryRun := args.DryRun

	paths := parsePaths(pathsStr)
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("at least one path is required")
	}

	// Ensure destination exists
	destFull := filepath.Join(v.path, destination)
	if !v.isPathSafe(destFull) {
		return nil, nil, fmt.Errorf("destination must be within vault")
	}
	if !dryRun {
		if err := os.MkdirAll(destFull, 0o755); err != nil {
			return nil, nil, fmt.Errorf("failed to create destination: %v", err)
		}
	}

	var results []string
	var errors []string

	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			p += ".md"
		}

		oldPath := filepath.Join(v.path, p)
		if !v.isPathSafe(oldPath) {
			errors = append(errors, fmt.Sprintf("%s: path must be within vault", p))
			continue
		}
		filename := filepath.Base(p)
		newPath := filepath.Join(destFull, filename)
		newRelPath := filepath.Join(destination, filename)
		if !v.isPathSafe(newPath) {
			errors = append(errors, fmt.Sprintf("%s: destination path must be within vault", p))
			continue
		}

		// Check source exists
		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("%s: not found", p))
			continue
		}

		// Check destination doesn't exist
		if _, err := os.Stat(newPath); err == nil {
			errors = append(errors, fmt.Sprintf("%s: already exists at destination", filename))
			continue
		}

		if !dryRun {
			// Move the file
			if err := os.Rename(oldPath, newPath); err != nil {
				errors = append(errors, fmt.Sprintf("%s: move failed", p))
				continue
			}

			// Update links if requested
			if updateLinks {
				oldName := strings.TrimSuffix(filename, ".md")
				_ = v.updateLinksInVault(oldName, oldName)
			}
		}
		if dryRun {
			results = append(results, fmt.Sprintf("%s -> %s (dry run)", p, newRelPath))
		} else {
			results = append(results, fmt.Sprintf("%s -> %s", p, newRelPath))
		}
	}

	var sb strings.Builder
	if dryRun {
		sb.WriteString(fmt.Sprintf("# Dry Run: Bulk Move to %s\n\n", destination))
	} else {
		sb.WriteString(fmt.Sprintf("# Bulk Move to %s\n\n", destination))
	}

	if len(results) > 0 {
		sb.WriteString("## Moved\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	if len(errors) > 0 {
		sb.WriteString("\n## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// BulkSetFrontmatterHandler sets a frontmatter property on multiple notes
func (v *Vault) BulkSetFrontmatterHandler(ctx context.Context, req *mcp.CallToolRequest, args BulkSetFrontmatterArgs) (*mcp.CallToolResult, any, error) {
	pathsStr := args.Paths
	key := args.Key
	value := args.Value
	dryRun := args.DryRun

	paths := parsePaths(pathsStr)
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("at least one path is required")
	}

	var results []string
	var errors []string

	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			p += ".md"
		}

		fullPath := filepath.Join(v.path, p)
		if !v.isPathSafe(fullPath) {
			errors = append(errors, fmt.Sprintf("%s: path must be within vault", p))
			continue
		}
		content, err := os.ReadFile(fullPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read failed", p))
			continue
		}

		contentStr := string(content)
		newContent := setFrontmatterKey(contentStr, key, value)

		if !dryRun {
			if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
				errors = append(errors, fmt.Sprintf("%s: write failed", p))
				continue
			}
		}
		if dryRun {
			results = append(results, fmt.Sprintf("%s: would set %s=%s", p, key, value))
		} else {
			results = append(results, fmt.Sprintf("%s: set %s=%s", p, key, value))
		}
	}

	var sb strings.Builder
	if dryRun {
		sb.WriteString(fmt.Sprintf("# Dry Run: Bulk Set Frontmatter: %s\n\n", key))
	} else {
		sb.WriteString(fmt.Sprintf("# Bulk Set Frontmatter: %s\n\n", key))
	}

	if len(results) > 0 {
		sb.WriteString("## Updated\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	if len(errors) > 0 {
		sb.WriteString("\n## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// parsePaths parses comma-separated list or JSON array of paths
func parsePaths(pathsStr string) []string {
	// Try parsing as JSON array first
	var paths []string
	if err := json.Unmarshal([]byte(pathsStr), &paths); err == nil {
		return paths
	}

	// Fallback to comma-separated string
	// Handle potential brackets if it looked like an array but wasn't valid JSON
	cleaned := strings.Trim(pathsStr, "[]")
	parts := strings.Split(cleaned, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
