package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// stubInfo holds information about a stub note
type stubInfo struct {
	path      string
	wordCount int
	modTime   time.Time
}

// FindStubsHandler finds notes with few words
func (v *Vault) FindStubsHandler(ctx context.Context, req *mcp.CallToolRequest, args FindStubsArgs) (*mcp.CallToolResult, any, error) {
	maxWords := args.MaxWords
	dir := args.Directory
	limit := args.Limit

	if maxWords <= 0 {
		maxWords = 100
	}
	if limit <= 0 {
		limit = 50
	}

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	var stubs []stubInfo

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		body := RemoveFrontmatter(string(content))
		wordCount := len(strings.Fields(body))

		if wordCount <= maxWords {
			relPath, _ := filepath.Rel(v.path, path)
			stubs = append(stubs, stubInfo{
				path:      relPath,
				wordCount: wordCount,
				modTime:   info.ModTime(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan vault: %v", err)
	}

	if len(stubs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No stub notes found (notes with ≤%d words)", maxWords)},
			},
		}, nil, nil
	}

	// Sort by word count ascending
	sort.Slice(stubs, func(i, j int) bool {
		return stubs[i].wordCount < stubs[j].wordCount
	})

	if limit > 0 && len(stubs) > limit {
		stubs = stubs[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Stub Notes (≤%d words)\n\n", maxWords)
	fmt.Fprintf(&sb, "Found %d notes that may need expansion:\n\n", len(stubs))

	for _, s := range stubs {
		fmt.Fprintf(&sb, "- **%s** (%d words) - last modified %s\n",
			s.path, s.wordCount, s.modTime.Format("Jan 2, 2006"))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// outdatedInfo holds information about an outdated note
type outdatedInfo struct {
	path      string
	modTime   time.Time
	daysSince int
}

// FindOutdatedHandler finds notes not modified in a while
func (v *Vault) FindOutdatedHandler(ctx context.Context, req *mcp.CallToolRequest, args FindOutdatedArgs) (*mcp.CallToolResult, any, error) {
	days := args.Days
	dir := args.Directory
	limit := args.Limit

	if days <= 0 {
		days = 90
	}
	if limit <= 0 {
		limit = 50
	}

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	var outdated []outdatedInfo

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		if info.ModTime().Before(cutoff) {
			relPath, _ := filepath.Rel(v.path, path)
			daysSince := int(time.Since(info.ModTime()).Hours() / 24)
			outdated = append(outdated, outdatedInfo{
				path:      relPath,
				modTime:   info.ModTime(),
				daysSince: daysSince,
			})
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan vault: %v", err)
	}

	if len(outdated) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No outdated notes found (all modified within %d days)", days)},
			},
		}, nil, nil
	}

	// Sort by oldest first
	sort.Slice(outdated, func(i, j int) bool {
		return outdated[i].modTime.Before(outdated[j].modTime)
	})

	if limit > 0 && len(outdated) > limit {
		outdated = outdated[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Outdated Notes (>%d days old)\n\n", days)
	fmt.Fprintf(&sb, "Found %d notes that haven't been touched:\n\n", len(outdated))

	for _, o := range outdated {
		fmt.Fprintf(&sb, "- **%s** - %d days ago (%s)\n",
			o.path, o.daysSince, o.modTime.Format("Jan 2, 2006"))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// unlinkedMention represents text that could be linked
type unlinkedMention struct {
	noteName string
	foundIn  string
	line     int
	context  string
}

// UnlinkedMentionsHandler finds text matching note names that aren't linked
func (v *Vault) UnlinkedMentionsHandler(ctx context.Context, req *mcp.CallToolRequest, args UnlinkedMentionsArgs) (*mcp.CallToolResult, any, error) {
	targetPath := args.Path

	if !strings.HasSuffix(targetPath, ".md") {
		targetPath += ".md"
	}

	fullPath := filepath.Join(v.path, targetPath)
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	// Get the note name to search for
	noteName := strings.TrimSuffix(filepath.Base(targetPath), ".md")
	noteNameLower := strings.ToLower(noteName)

	var mentions []unlinkedMention

	err := filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip the target note itself
		if path == fullPath {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		lines := strings.Split(string(content), "\n")

		for i, line := range lines {
			lineLower := strings.ToLower(line)

			// Check if the note name appears in this line
			if !strings.Contains(lineLower, noteNameLower) {
				continue
			}

			// Check if it's already linked
			if isAlreadyLinked(line, noteName) {
				continue
			}

			// Found an unlinked mention
			ctxLine := line
			if len(ctxLine) > 100 {
				ctxLine = ctxLine[:100] + "..."
			}

			mentions = append(mentions, unlinkedMention{
				noteName: noteName,
				foundIn:  relPath,
				line:     i + 1,
				context:  strings.TrimSpace(ctxLine),
			})
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan vault: %v", err)
	}

	if len(mentions) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No unlinked mentions of '%s' found", noteName)},
			},
		}, nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Unlinked Mentions of '%s'\n\n", noteName)
	fmt.Fprintf(&sb, "Found %d places where '%s' appears but isn't linked:\n\n", len(mentions), noteName)

	// Group by file
	byFile := make(map[string][]unlinkedMention)
	for _, m := range mentions {
		byFile[m.foundIn] = append(byFile[m.foundIn], m)
	}

	for file, ms := range byFile {
		fmt.Fprintf(&sb, "## %s\n", file)
		for _, m := range ms {
			fmt.Fprintf(&sb, "- L%d: %s\n", m.line, m.context)
		}
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// linkSuggestion represents a suggested link
type linkSuggestion struct {
	targetNote string
	reason     string
	strength   int // higher = stronger suggestion
}

// SuggestLinksHandler suggests notes that should be linked
func (v *Vault) SuggestLinksHandler(ctx context.Context, req *mcp.CallToolRequest, args SuggestLinksArgs) (*mcp.CallToolResult, any, error) {
	notePath := args.Path
	limit := args.Limit

	if limit <= 0 {
		limit = 10
	}

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

	body := RemoveFrontmatter(string(content))
	bodyLower := strings.ToLower(body)
	existingLinks := ExtractWikilinks(body)
	existingSet := make(map[string]bool)
	for _, l := range existingLinks {
		existingSet[strings.ToLower(l)] = true
	}

	suggestions := make(map[string]*linkSuggestion)

	// Scan all notes and find potential links
	err = filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		if path == fullPath {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		otherName := strings.TrimSuffix(filepath.Base(path), ".md")
		otherNameLower := strings.ToLower(otherName)

		// Skip if already linked
		if existingSet[otherNameLower] {
			return nil
		}

		// Check if the other note's name appears in our content
		if strings.Contains(bodyLower, otherNameLower) && len(otherName) > 2 {
			count := strings.Count(bodyLower, otherNameLower)
			suggestions[relPath] = &linkSuggestion{
				targetNote: relPath,
				reason:     fmt.Sprintf("'%s' mentioned %d time(s)", otherName, count),
				strength:   count * 10,
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan vault: %v", err)
	}

	if len(suggestions) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No link suggestions for: %s", notePath)},
			},
		}, nil, nil
	}

	// Sort by strength
	var sorted []linkSuggestion
	for _, s := range suggestions {
		sorted = append(sorted, *s)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].strength > sorted[j].strength
	})

	if limit > 0 && len(sorted) > limit {
		sorted = sorted[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Link Suggestions for %s\n\n", notePath)
	fmt.Fprintf(&sb, "Notes that could be linked:\n\n")

	for _, s := range sorted {
		fmt.Fprintf(&sb, "- [[%s]] - %s\n",
			strings.TrimSuffix(s.targetNote, ".md"), s.reason)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// isAlreadyLinked checks if text contains a wikilink to the given note
func isAlreadyLinked(line, noteName string) bool {
	// Check for [[noteName]] or [[noteName|...]]
	patterns := []string{
		"[[" + noteName + "]]",
		"[[" + noteName + "|",
		"[[" + strings.ToLower(noteName) + "]]",
		"[[" + strings.ToLower(noteName) + "|",
	}

	lineLower := strings.ToLower(line)
	for _, p := range patterns {
		if strings.Contains(lineLower, strings.ToLower(p)) {
			return true
		}
	}

	return false
}
