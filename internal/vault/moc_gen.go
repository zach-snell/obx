package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// noteInfo holds metadata about a note for MOC/index generation
type noteInfo struct {
	path     string
	name     string
	title    string
	tags     []string
	hasLinks bool
}

// collectNotes gathers note info from a directory
func (v *Vault) collectNotes(dir string, recursive bool) ([]noteInfo, error) {
	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var notes []noteInfo

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories unless we're being recursive
		if info.IsDir() {
			if path != searchPath && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		relPath, _ := filepath.Rel(v.path, path)
		name := strings.TrimSuffix(filepath.Base(path), ".md")

		title := ExtractH1Title(contentStr)
		if title == "" {
			title = name
		}

		notes = append(notes, noteInfo{
			path:     relPath,
			name:     name,
			title:    title,
			tags:     ExtractTags(contentStr),
			hasLinks: len(ExtractWikilinks(contentStr)) > 0,
		})

		return nil
	}

	if err := filepath.Walk(searchPath, walkFn); err != nil {
		return nil, err
	}

	return notes, nil
}

// writeGeneratedFile writes content to a file, ensuring directory exists
func (v *Vault) writeGeneratedFile(output, content, fileType string, noteCount int) (*mcp.CallToolResult, error) {
	if !strings.HasSuffix(output, ".md") {
		output += ".md"
	}
	fullPath := filepath.Join(v.path, output)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %v", fileType, err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Generated %s at %s with %d notes", fileType, output, noteCount)},
		},
	}, nil
}

// GenerateMOCHandler generates a new Map of Content from a directory
func (v *Vault) GenerateMOCHandler(ctx context.Context, req *mcp.CallToolRequest, args GenerateMOCArgs) (*mcp.CallToolResult, any, error) {
	dir := args.Directory
	title := args.Title
	output := args.Output
	groupBy := args.GroupBy // none, tag, alpha
	recursive := args.Recursive

	if title == "" {
		if dir != "" {
			title = filepath.Base(dir) + " MOC"
		} else {
			title = "Vault MOC"
		}
	}
	if groupBy == "" {
		groupBy = "none"
	}

	notes, err := v.collectNotes(dir, recursive)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect notes: %v", err)
	}

	if len(notes) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No notes found in directory"},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---\ntags: [moc]\n---\n\n# %s\n\n", title))

	switch groupBy {
	case "alpha":
		sb.WriteString(formatByAlpha(notes))
	case "tag":
		sb.WriteString(formatByTag(notes))
	default:
		sb.WriteString(formatFlat(notes))
	}

	content := sb.String()

	if output != "" {
		res, err := v.writeGeneratedFile(output, content, "MOC", len(notes))
		if err != nil {
			return nil, nil, err
		}
		return res, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: content},
		},
	}, nil, nil
}

// formatFlat creates a simple flat list of notes
func formatFlat(notes []noteInfo) string {
	sort.Slice(notes, func(i, j int) bool {
		return strings.ToLower(notes[i].title) < strings.ToLower(notes[j].title)
	})

	var sb strings.Builder
	for _, n := range notes {
		sb.WriteString(fmt.Sprintf("- [[%s]]\n", n.name))
	}
	return sb.String()
}

// formatByAlpha groups notes alphabetically
func formatByAlpha(notes []noteInfo) string {
	sort.Slice(notes, func(i, j int) bool {
		return strings.ToLower(notes[i].title) < strings.ToLower(notes[j].title)
	})

	groups := make(map[rune][]noteInfo)
	for _, n := range notes {
		if len(n.title) == 0 {
			continue
		}
		firstRune := unicode.ToUpper(rune(n.title[0]))
		if !unicode.IsLetter(firstRune) {
			firstRune = '#'
		}
		groups[firstRune] = append(groups[firstRune], n)
	}

	var keys []rune
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i] == '#' {
			return false
		}
		if keys[j] == '#' {
			return true
		}
		return keys[i] < keys[j]
	})

	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(fmt.Sprintf("## %c\n\n", key))
		for _, n := range groups[key] {
			sb.WriteString(fmt.Sprintf("- [[%s]]\n", n.name))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// formatByTag groups notes by their primary tag
func formatByTag(notes []noteInfo) string {
	groups := make(map[string][]noteInfo)
	for _, n := range notes {
		tag := "Untagged"
		if len(n.tags) > 0 {
			tag = n.tags[0]
		}
		groups[tag] = append(groups[tag], n)
	}

	var keys []string
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(fmt.Sprintf("## %s\n\n", key))
		groupNotes := groups[key]
		sort.Slice(groupNotes, func(i, j int) bool {
			return strings.ToLower(groupNotes[i].title) < strings.ToLower(groupNotes[j].title)
		})
		for _, n := range groupNotes {
			sb.WriteString(fmt.Sprintf("- [[%s]]\n", n.name))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// UpdateMOCHandler updates an existing MOC with new notes
func (v *Vault) UpdateMOCHandler(ctx context.Context, req *mcp.CallToolRequest, args UpdateMOCArgs) (*mcp.CallToolResult, any, error) {
	mocPath := args.Path
	dir := args.Directory
	recursive := args.Recursive

	if !strings.HasSuffix(mocPath, ".md") {
		mocPath += ".md"
	}

	fullPath := filepath.Join(v.path, mocPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read MOC: %v", err)
	}

	existingLinks := ExtractWikilinks(string(content))
	existingSet := make(map[string]bool)
	for _, link := range existingLinks {
		existingSet[strings.ToLower(link)] = true
	}

	notes, err := v.collectNotes(dir, recursive)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect notes: %v", err)
	}

	var newNotes []noteInfo
	for _, n := range notes {
		// Skip the MOC itself
		if n.path == mocPath {
			continue
		}
		if !existingSet[strings.ToLower(n.name)] {
			newNotes = append(newNotes, n)
		}
	}

	if len(newNotes) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "MOC is up to date, no new notes found"},
			},
		}, nil, nil
	}

	sort.Slice(newNotes, func(i, j int) bool {
		return strings.ToLower(newNotes[i].title) < strings.ToLower(newNotes[j].title)
	})

	var sb strings.Builder
	sb.WriteString("\n\n## New Notes\n\n")
	for _, n := range newNotes {
		sb.WriteString(fmt.Sprintf("- [[%s]]\n", n.name))
	}

	updatedContent := string(content) + sb.String()
	if err := os.WriteFile(fullPath, []byte(updatedContent), 0o644); err != nil {
		return nil, nil, fmt.Errorf("failed to update MOC: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Added %d new notes to %s", len(newNotes), mocPath)},
		},
	}, nil, nil
}

// GenerateIndexHandler generates an alphabetical index of all notes
func (v *Vault) GenerateIndexHandler(ctx context.Context, req *mcp.CallToolRequest, args GenerateIndexArgs) (*mcp.CallToolResult, any, error) {
	dir := args.Directory
	output := args.Output
	title := args.Title
	includeOrphans := args.IncludeOrphans

	if title == "" {
		title = "Index"
	}

	notes, err := v.collectNotes(dir, true)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect notes: %v", err)
	}

	if len(notes) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No notes found"},
			},
		}, nil, nil
	}

	// Filter orphans if requested
	if !includeOrphans {
		var filtered []noteInfo
		for _, n := range notes {
			if n.hasLinks || len(n.tags) > 0 {
				filtered = append(filtered, n)
			}
		}
		notes = filtered
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---\ntags: [index, moc]\n---\n\n# %s\n\n", title))
	sb.WriteString(fmt.Sprintf("Total: %d notes\n\n", len(notes)))
	sb.WriteString(formatByAlpha(notes))

	content := sb.String()

	if output != "" {
		res, err := v.writeGeneratedFile(output, content, "index", len(notes))
		if err != nil {
			return nil, nil, err
		}
		return res, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: content},
		},
	}, nil, nil
}
