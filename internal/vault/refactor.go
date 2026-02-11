package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// SplitNoteHandler splits a note at a heading into separate notes
func (v *Vault) SplitNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	level := int(req.GetInt("level", 2))
	keepOriginal := req.GetBool("keep_original", false)
	outputDir := req.GetString("output_dir", "")

	if !strings.HasSuffix(path, ".md") {
		path += ".md"
	}

	fullPath := filepath.Join(v.path, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	// Parse into sections by heading level
	sections := splitByHeading(string(content), level)

	if len(sections) <= 1 {
		return mcp.NewToolResultText("No sections found to split at the specified heading level"), nil
	}

	// Determine output directory
	if outputDir == "" {
		outputDir = filepath.Dir(path)
	}
	outputDirFull := filepath.Join(v.path, outputDir)
	if err := os.MkdirAll(outputDirFull, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create output directory: %v", err)), nil
	}

	var created []string
	for _, section := range sections {
		if section.title == "" {
			continue // Skip preamble without title
		}

		// Create sanitized filename from title
		filename := sanitizeFilename(section.title) + ".md"
		newPath := filepath.Join(outputDirFull, filename)

		// Add frontmatter if extracting
		newContent := fmt.Sprintf("# %s\n\n%s", section.title, strings.TrimSpace(section.content))

		//#nosec G306 -- Obsidian notes need to be readable by the user
		if err := os.WriteFile(newPath, []byte(newContent), 0o644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write %s: %v", filename, err)), nil
		}

		relPath, _ := filepath.Rel(v.path, newPath)
		created = append(created, relPath)
	}

	if !keepOriginal {
		if err := os.Remove(fullPath); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to remove original: %v", err)), nil
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Split into %d notes:\n\n", len(created)))
	for _, c := range created {
		sb.WriteString(fmt.Sprintf("- %s\n", c))
	}
	if !keepOriginal {
		sb.WriteString(fmt.Sprintf("\nOriginal note removed: %s", path))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// section represents a heading section
type section struct {
	title   string
	content string
}

// splitByHeading splits content by heading level
func splitByHeading(content string, level int) []section {
	lines := strings.Split(content, "\n")
	prefix := strings.Repeat("#", level) + " "

	var sections []section
	var current section
	var contentLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, prefix) && !strings.HasPrefix(line, prefix+"#") {
			// Save previous section
			if current.title != "" || len(contentLines) > 0 {
				current.content = strings.Join(contentLines, "\n")
				sections = append(sections, current)
			}
			// Start new section
			current = section{title: strings.TrimPrefix(line, prefix)}
			contentLines = nil
		} else {
			contentLines = append(contentLines, line)
		}
	}

	// Save last section
	if current.title != "" || len(contentLines) > 0 {
		current.content = strings.Join(contentLines, "\n")
		sections = append(sections, current)
	}

	return sections
}

// sanitizeFilename creates a safe filename from a string
func sanitizeFilename(s string) string {
	// Replace problematic characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	result := replacer.Replace(s)
	result = strings.TrimSpace(result)
	if result == "" {
		result = "untitled"
	}
	return result
}

// MergeNotesHandler merges multiple notes into one
func (v *Vault) MergeNotesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathsStr, err := req.RequireString("paths")
	if err != nil {
		return mcp.NewToolResultError("paths is required"), nil
	}

	output, err := req.RequireString("output")
	if err != nil {
		return mcp.NewToolResultError("output path is required"), nil
	}

	separator := req.GetString("separator", "\n\n---\n\n")
	deleteOriginals := req.GetBool("delete_originals", false)
	addHeadings := req.GetBool("add_headings", true)

	// Parse paths (comma-separated or JSON array)
	paths := parsePaths(pathsStr)
	if len(paths) < 2 {
		return mcp.NewToolResultError("at least 2 paths are required to merge"), nil
	}

	var contents []string
	var validPaths []string

	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			p += ".md"
		}

		fullPath := filepath.Join(v.path, p)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read %s: %v", p, err)), nil
		}

		contentStr := string(content)

		if addHeadings {
			// Add the filename as a heading if it doesn't start with one
			name := strings.TrimSuffix(filepath.Base(p), ".md")
			if !strings.HasPrefix(strings.TrimSpace(contentStr), "#") {
				contentStr = fmt.Sprintf("## %s\n\n%s", name, contentStr)
			}
		}

		contents = append(contents, strings.TrimSpace(contentStr))
		validPaths = append(validPaths, p)
	}

	// Combine contents
	merged := strings.Join(contents, separator)

	// Write output
	if !strings.HasSuffix(output, ".md") {
		output += ".md"
	}
	outputFull := filepath.Join(v.path, output)
	if err := os.MkdirAll(filepath.Dir(outputFull), 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	//#nosec G306 -- Obsidian notes need to be readable by the user
	if err := os.WriteFile(outputFull, []byte(merged), 0o644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write merged note: %v", err)), nil
	}

	// Delete originals if requested
	if deleteOriginals {
		for _, p := range validPaths {
			fullPath := filepath.Join(v.path, p)
			if err := os.Remove(fullPath); err != nil {
				// Log but don't fail
				continue
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Merged %d notes into %s\n\n", len(validPaths), output))
	sb.WriteString("Source notes:\n")
	for _, p := range validPaths {
		if deleteOriginals {
			sb.WriteString(fmt.Sprintf("- %s (deleted)\n", p))
		} else {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// parsePaths parses a comma-separated or JSON array of paths
func parsePaths(pathsStr string) []string {
	pathsStr = strings.TrimSpace(pathsStr)

	// Try JSON array first
	if strings.HasPrefix(pathsStr, "[") {
		// Simple JSON array parsing
		pathsStr = strings.TrimPrefix(pathsStr, "[")
		pathsStr = strings.TrimSuffix(pathsStr, "]")
		var paths []string
		for _, p := range strings.Split(pathsStr, ",") {
			p = strings.TrimSpace(p)
			p = strings.Trim(p, "\"'")
			if p != "" {
				paths = append(paths, p)
			}
		}
		return paths
	}

	// Comma-separated
	var paths []string
	for _, p := range strings.Split(pathsStr, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// ExtractSectionHandler extracts a section to a new note
func (v *Vault) ExtractSectionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	heading, err := req.RequireString("heading")
	if err != nil {
		return mcp.NewToolResultError("heading is required"), nil
	}

	output := req.GetString("output", "")
	removeFromOriginal := req.GetBool("remove_from_original", true)
	addLink := req.GetBool("add_link", true)

	if !strings.HasSuffix(path, ".md") {
		path += ".md"
	}

	fullPath := filepath.Join(v.path, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	// Extract the section
	sectionContent := extractSection(string(content), heading)
	if sectionContent == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Heading '%s' not found", heading)), nil
	}

	// Determine output path
	if output == "" {
		output = sanitizeFilename(heading) + ".md"
	}
	if !strings.HasSuffix(output, ".md") {
		output += ".md"
	}

	outputFull := filepath.Join(v.path, output)
	if err := os.MkdirAll(filepath.Dir(outputFull), 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	// Create new note with extracted content
	newContent := fmt.Sprintf("# %s\n\n%s", heading, strings.TrimSpace(sectionContent))

	//#nosec G306 -- Obsidian notes need to be readable by the user
	if err := os.WriteFile(outputFull, []byte(newContent), 0o644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write new note: %v", err)), nil
	}

	// Modify original if requested
	if removeFromOriginal {
		contentStr := string(content)
		newOriginal := removeSectionFromContent(contentStr, heading)

		if addLink {
			// Add link to the extracted note
			noteName := strings.TrimSuffix(filepath.Base(output), ".md")
			linkText := fmt.Sprintf("\n\nSee: [[%s]]\n", noteName)
			newOriginal += linkText
		}

		//#nosec G306 -- Obsidian notes need to be readable by the user
		if err := os.WriteFile(fullPath, []byte(newOriginal), 0o644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update original: %v", err)), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Extracted section '%s' to %s", heading, output)), nil
}

// removeSectionFromContent removes a section from content
func removeSectionFromContent(content, heading string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inSection := false
	sectionLevel := 0

	for _, line := range lines {
		// Check if this is a heading
		if strings.HasPrefix(line, "#") {
			level := 0
			for _, c := range line {
				if c == '#' {
					level++
				} else {
					break
				}
			}

			headingText := strings.TrimSpace(strings.TrimLeft(line, "#"))

			if strings.EqualFold(headingText, heading) {
				inSection = true
				sectionLevel = level
				continue
			}

			if inSection && level <= sectionLevel {
				inSection = false
			}
		}

		if !inSection {
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// DuplicateNoteHandler duplicates a note
func (v *Vault) DuplicateNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	output := req.GetString("output", "")

	if !strings.HasSuffix(path, ".md") {
		path += ".md"
	}

	fullPath := filepath.Join(v.path, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	// Determine output path
	if output == "" {
		baseName := strings.TrimSuffix(filepath.Base(path), ".md")
		dir := filepath.Dir(path)
		output = filepath.Join(dir, baseName+" (copy).md")
	}
	if !strings.HasSuffix(output, ".md") {
		output += ".md"
	}

	outputFull := filepath.Join(v.path, output)

	// Check if output already exists
	if _, err := os.Stat(outputFull); err == nil {
		return mcp.NewToolResultError(fmt.Sprintf("File already exists: %s", output)), nil
	}

	if err := os.MkdirAll(filepath.Dir(outputFull), 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	//#nosec G306 -- Obsidian notes need to be readable by the user
	if err := os.WriteFile(outputFull, content, 0o644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write duplicate: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Duplicated %s to %s", path, output)), nil
}
