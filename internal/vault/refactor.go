package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ExtractNoteHandler splits a note at a heading into separate notes
func (v *Vault) ExtractNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args ExtractNoteArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	level := args.Level
	keepOriginal := args.KeepOriginal
	outputDir := args.OutputDir
	dryRun := args.DryRun

	if level <= 0 {
		level = 2
	}

	if !strings.HasSuffix(path, ".md") {
		path += ".md"
	}

	fullPath := filepath.Join(v.path, path)

	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}

	// Parse into sections by heading level
	sections := splitByHeading(string(content), level)

	if len(sections) <= 1 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No sections found to split at the specified heading level"},
			},
		}, nil, nil
	}

	// Determine output directory
	if outputDir == "" {
		outputDir = filepath.Dir(path)
	}
	outputDirFull := filepath.Join(v.path, outputDir)
	if !v.isPathSafe(outputDirFull) {
		return nil, nil, fmt.Errorf("output directory must be within vault")
	}

	if !dryRun {
		if err := os.MkdirAll(outputDirFull, 0o755); err != nil {
			return nil, nil, fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	created, err := v.writeSplitSections(sections, outputDirFull, dryRun)
	if err != nil {
		return nil, nil, err
	}

	if !keepOriginal {
		if !dryRun {
			if err := os.Remove(fullPath); err != nil {
				return nil, nil, fmt.Errorf("failed to remove original: %v", err)
			}
		}
	}

	var sb strings.Builder
	if dryRun {
		sb.WriteString(fmt.Sprintf("Dry run: would split into %d notes:\n\n", len(created)))
	} else {
		sb.WriteString(fmt.Sprintf("Split into %d notes:\n\n", len(created)))
	}
	for _, c := range created {
		sb.WriteString(fmt.Sprintf("- %s\n", c))
	}
	if !keepOriginal {
		if dryRun {
			sb.WriteString(fmt.Sprintf("\nOriginal note would be removed: %s", path))
		} else {
			sb.WriteString(fmt.Sprintf("\nOriginal note removed: %s", path))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// writeSplitSections writes each titled section to its own file and returns the created paths.
func (v *Vault) writeSplitSections(sections []section, outputDir string, dryRun bool) ([]string, error) {
	var created []string
	for _, sec := range sections {
		if sec.title == "" {
			continue
		}
		filename := sanitizeFilename(sec.title) + ".md"
		newPath := filepath.Join(outputDir, filename)
		newContent := fmt.Sprintf("# %s\n\n%s", sec.title, strings.TrimSpace(sec.content))
		if !dryRun {
			if err := os.WriteFile(newPath, []byte(newContent), 0o600); err != nil {
				return nil, fmt.Errorf("failed to write %s: %v", filename, err)
			}
		}
		relPath, _ := filepath.Rel(v.path, newPath)
		created = append(created, relPath)
	}
	return created, nil
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
func (v *Vault) MergeNotesHandler(ctx context.Context, req *mcp.CallToolRequest, args MergeNotesArgs) (*mcp.CallToolResult, any, error) {
	pathsStr := args.Paths
	output := args.Output
	separator := args.Separator
	deleteOriginals := args.DeleteOriginals
	addHeadings := args.AddHeadings
	dryRun := args.DryRun

	if separator == "" {
		separator = "\n\n---\n\n"
	}

	// Parse paths (comma-separated or JSON array) - using parsePaths from bulk.go
	paths := parsePaths(pathsStr)
	if len(paths) < 2 {
		return nil, nil, fmt.Errorf("at least 2 paths are required to merge")
	}

	contents, validPaths, err := v.readMergeContents(paths, addHeadings)
	if err != nil {
		return nil, nil, err
	}

	// Combine contents
	merged := strings.Join(contents, separator)

	// Write output
	if !strings.HasSuffix(output, ".md") {
		output += ".md"
	}
	outputFull := filepath.Join(v.path, output)
	if !v.isPathSafe(outputFull) {
		return nil, nil, fmt.Errorf("output path must be within vault")
	}

	if !dryRun {
		if err := os.MkdirAll(filepath.Dir(outputFull), 0o755); err != nil {
			return nil, nil, fmt.Errorf("failed to create directory: %v", err)
		}
	}

	if !dryRun {
		if err := os.WriteFile(outputFull, []byte(merged), 0o600); err != nil {
			return nil, nil, fmt.Errorf("failed to write merged note: %v", err)
		}
	}

	// Delete originals if requested
	if deleteOriginals && !dryRun {
		for _, p := range validPaths {
			fullPath := filepath.Join(v.path, p)
			if err := os.Remove(fullPath); err != nil {
				// Log but don't fail
				continue
			}
		}
	}

	var sb strings.Builder
	if dryRun {
		sb.WriteString(fmt.Sprintf("Dry run: would merge %d notes into %s\n\n", len(validPaths), output))
	} else {
		sb.WriteString(fmt.Sprintf("Merged %d notes into %s\n\n", len(validPaths), output))
	}
	sb.WriteString("Source notes:\n")
	for _, p := range validPaths {
		if deleteOriginals {
			if dryRun {
				sb.WriteString(fmt.Sprintf("- %s (would be deleted)\n", p))
			} else {
				sb.WriteString(fmt.Sprintf("- %s (deleted)\n", p))
			}
		} else {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// readMergeContents reads and prepares note contents for merging.
func (v *Vault) readMergeContents(paths []string, addHeadings bool) (contents, validPaths []string, err error) {
	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			p += ".md"
		}
		fullPath := filepath.Join(v.path, p)
		if !v.isPathSafe(fullPath) {
			return nil, nil, fmt.Errorf("path must be within vault: %s", p)
		}
		data, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			return nil, nil, fmt.Errorf("failed to read %s: %v", p, readErr)
		}
		contentStr := string(data)
		if addHeadings {
			name := strings.TrimSuffix(filepath.Base(p), ".md")
			if !strings.HasPrefix(strings.TrimSpace(contentStr), "#") {
				contentStr = fmt.Sprintf("## %s\n\n%s", name, contentStr)
			}
		}
		contents = append(contents, strings.TrimSpace(contentStr))
		validPaths = append(validPaths, p)
	}
	return contents, validPaths, nil
}

// ExtractSectionHandler extracts a section to a new note
func (v *Vault) ExtractSectionHandler(ctx context.Context, req *mcp.CallToolRequest, args ExtractSectionArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	heading := args.Heading
	output := args.Output
	removeFromOriginal := args.RemoveFromOriginal
	addLink := args.AddLink
	dryRun := args.DryRun

	if !strings.HasSuffix(path, ".md") {
		path += ".md"
	}

	fullPath := filepath.Join(v.path, path)
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}

	// Extract the section (using extractSection from batch.go)
	sectionContent := extractSection(string(content), heading)
	if sectionContent == "" {
		return nil, nil, fmt.Errorf("heading '%s' not found", heading)
	}

	// Determine output path
	if output == "" {
		output = sanitizeFilename(heading) + ".md"
	}
	if !strings.HasSuffix(output, ".md") {
		output += ".md"
	}

	outputFull := filepath.Join(v.path, output)
	if !v.isPathSafe(outputFull) {
		return nil, nil, fmt.Errorf("output path must be within vault")
	}

	if !dryRun {
		if err := os.MkdirAll(filepath.Dir(outputFull), 0o755); err != nil {
			return nil, nil, fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// Create new note with extracted content
	newContent := fmt.Sprintf("# %s\n\n%s", heading, strings.TrimSpace(sectionContent))

	if !dryRun {
		if err := os.WriteFile(outputFull, []byte(newContent), 0o600); err != nil {
			return nil, nil, fmt.Errorf("failed to write new note: %v", err)
		}
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

		if !dryRun {
			if err := os.WriteFile(fullPath, []byte(newOriginal), 0o600); err != nil {
				return nil, nil, fmt.Errorf("failed to update original: %v", err)
			}
		}
	}

	resultMsg := fmt.Sprintf("Extracted section '%s' to %s", heading, output)
	if dryRun {
		resultMsg = fmt.Sprintf("Dry run: would extract section '%s' to %s", heading, output)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: resultMsg},
		},
	}, nil, nil
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
func (v *Vault) DuplicateNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args DuplicateNoteArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	output := args.Output

	if !strings.HasSuffix(path, ".md") {
		path += ".md"
	}

	fullPath := filepath.Join(v.path, path)

	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
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
	if !v.isPathSafe(outputFull) {
		return nil, nil, fmt.Errorf("output path must be within vault")
	}

	// Check if output already exists
	if _, err := os.Stat(outputFull); err == nil {
		return nil, nil, fmt.Errorf("file already exists: %s", output)
	}

	if err := os.MkdirAll(filepath.Dir(outputFull), 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(outputFull, content, 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write duplicate: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Duplicated %s to %s", path, output)},
		},
	}, nil, nil
}
