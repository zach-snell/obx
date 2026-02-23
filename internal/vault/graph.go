package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ForwardLinksHandler shows outgoing links from a note
func (v *Vault) ForwardLinksHandler(ctx context.Context, req *mcp.CallToolRequest, args ForwardLinksArgs) (*mcp.CallToolResult, any, error) {
	notePath := args.Path

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.GetPath(), notePath)
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

	links := ExtractWikilinks(string(content))

	if len(links) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No outgoing links found in: %s", notePath)},
			},
		}, nil, nil
	}

	// Check which links exist
	var existing, broken []string
	for _, link := range links {
		if v.noteExists(link) {
			existing = append(existing, link)
		} else {
			broken = append(broken, link)
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Forward Links from %s\n\n", notePath)
	fmt.Fprintf(&sb, "Total: %d links (%d existing, %d broken)\n\n", len(links), len(existing), len(broken))

	if len(existing) > 0 {
		sb.WriteString("## Existing Notes\n")
		for _, link := range existing {
			fmt.Fprintf(&sb, "- [[%s]]\n", link)
		}
		sb.WriteString("\n")
	}

	if len(broken) > 0 {
		sb.WriteString("## Broken Links (no matching note)\n")
		for _, link := range broken {
			fmt.Fprintf(&sb, "- [[%s]] ⚠️\n", link)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// noteLinks represents a note and its link relationships
type noteLinks struct {
	path     string
	outgoing []string
	incoming int
}

// linkGraph represents the full vault link structure
type linkGraph struct {
	notes map[string]*noteLinks
}

// buildLinkGraph scans the vault and builds the link graph
func (v *Vault) buildLinkGraph(searchPath string) (*linkGraph, error) {
	graph := &linkGraph{notes: make(map[string]*noteLinks)}

	// Collect all notes and their outgoing links
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(v.GetPath(), path)
		noteName := strings.TrimSuffix(relPath, ".md")

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		graph.notes[noteName] = &noteLinks{
			path:     relPath,
			outgoing: ExtractWikilinks(string(content)),
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Count incoming links
	for _, note := range graph.notes {
		for _, link := range note.outgoing {
			graph.incrementIncoming(link)
		}
	}

	return graph, nil
}

// incrementIncoming increments the incoming count for a target note
func (g *linkGraph) incrementIncoming(link string) {
	target := normalizeNoteName(link)
	if targetNote, exists := g.notes[target]; exists {
		targetNote.incoming++
		return
	}
	// Try basename match
	for name, n := range g.notes {
		if filepath.Base(name) == link || name == link {
			n.incoming++
			return
		}
	}
}

// orphanResult holds categorized orphan notes
type orphanResult struct {
	trueOrphans []string // No incoming AND no outgoing
	noIncoming  []string // No incoming but has outgoing
	deadEnds    []string // Incoming but no outgoing
}

// findOrphans categorizes notes by their link status
func (g *linkGraph) findOrphans(includeDeadEnds bool) orphanResult {
	var result orphanResult

	for name, note := range g.notes {
		hasIncoming := note.incoming > 0
		hasOutgoing := len(note.outgoing) > 0

		switch {
		case !hasIncoming && !hasOutgoing:
			result.trueOrphans = append(result.trueOrphans, name)
		case !hasIncoming && hasOutgoing:
			result.noIncoming = append(result.noIncoming, name)
		case hasIncoming && !hasOutgoing:
			if includeDeadEnds {
				result.deadEnds = append(result.deadEnds, name)
			}
		}
	}

	sort.Strings(result.trueOrphans)
	sort.Strings(result.noIncoming)
	sort.Strings(result.deadEnds)

	return result
}

// formatOrphanResult formats the orphan analysis as markdown
func (g *linkGraph) formatOrphanResult(result orphanResult, includeDeadEnds bool) string {
	var sb strings.Builder
	sb.WriteString("# Orphan Analysis\n\n")

	if len(result.trueOrphans) > 0 {
		fmt.Fprintf(&sb, "## True Orphans (%d)\n", len(result.trueOrphans))
		sb.WriteString("Notes with no incoming or outgoing links:\n\n")
		for _, name := range result.trueOrphans {
			fmt.Fprintf(&sb, "- [[%s]]\n", name)
		}
		sb.WriteString("\n")
	}

	if len(result.noIncoming) > 0 {
		fmt.Fprintf(&sb, "## Unlinked Notes (%d)\n", len(result.noIncoming))
		sb.WriteString("Notes that link out but nothing links to them:\n\n")
		for _, name := range result.noIncoming {
			fmt.Fprintf(&sb, "- [[%s]] → %d outgoing\n", name, len(g.notes[name].outgoing))
		}
		sb.WriteString("\n")
	}

	if includeDeadEnds && len(result.deadEnds) > 0 {
		fmt.Fprintf(&sb, "## Dead Ends (%d)\n", len(result.deadEnds))
		sb.WriteString("Notes with incoming links but no outgoing:\n\n")
		for _, name := range result.deadEnds {
			fmt.Fprintf(&sb, "- [[%s]] ← %d incoming\n", name, g.notes[name].incoming)
		}
	}

	if len(result.trueOrphans) == 0 && len(result.noIncoming) == 0 {
		sb.WriteString("No orphan notes found! Your vault is well-connected.\n")
	}

	return sb.String()
}

// OrphanNotesHandler finds notes with no incoming or outgoing links
func (v *Vault) OrphanNotesHandler(ctx context.Context, req *mcp.CallToolRequest, args OrphanNotesArgs) (*mcp.CallToolResult, any, error) {
	dir := args.Directory
	includeDeadEnds := args.IncludeDeadEnds

	searchPath := v.GetPath()
	if dir != "" {
		searchPath = filepath.Join(v.GetPath(), dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	graph, err := v.buildLinkGraph(searchPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan vault: %v", err)
	}

	result := graph.findOrphans(includeDeadEnds)
	output := graph.formatOrphanResult(result, includeDeadEnds)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: output},
		},
	}, nil, nil
}

// brokenLink represents a wikilink that doesn't resolve
type brokenLink struct {
	source string
	target string
	line   int
}

// buildExistingNotesSet creates a set of all existing note names
func (v *Vault) buildExistingNotesSet() (map[string]bool, error) {
	existing := make(map[string]bool)
	err := filepath.Walk(v.GetPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		relPath, _ := filepath.Rel(v.GetPath(), path)
		noteName := strings.TrimSuffix(relPath, ".md")
		existing[noteName] = true
		existing[filepath.Base(noteName)] = true
		return nil
	})
	return existing, err
}

// findBrokenLinksInNote finds broken links in a single note
func findBrokenLinksInNote(relPath, content string, existing map[string]bool) []brokenLink {
	var broken []brokenLink
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		matches := wikilinkRegex.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			link := strings.TrimSpace(match[1])
			if link == "" || strings.HasPrefix(link, "http") || strings.Contains(link, "#") {
				continue
			}
			normalized := normalizeNoteName(link)
			if !existing[normalized] && !existing[link] {
				broken = append(broken, brokenLink{source: relPath, target: link, line: i + 1})
			}
		}
	}
	return broken
}

// formatBrokenLinks formats broken links as markdown
func formatBrokenLinks(broken []brokenLink) string {
	if len(broken) == 0 {
		return "No broken links found! All wikilinks resolve to existing notes."
	}

	// Group by source
	bySource := make(map[string][]brokenLink)
	for _, bl := range broken {
		bySource[bl.source] = append(bySource[bl.source], bl)
	}

	var sources []string
	for s := range bySource {
		sources = append(sources, s)
	}
	sort.Strings(sources)

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Broken Links (%d total in %d files)\n\n", len(broken), len(sources))

	for _, source := range sources {
		links := bySource[source]
		fmt.Fprintf(&sb, "## %s (%d broken)\n", source, len(links))
		for _, bl := range links {
			fmt.Fprintf(&sb, "- L%d: [[%s]]\n", bl.line, bl.target)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// BrokenLinksHandler finds wikilinks pointing to non-existent notes
func (v *Vault) BrokenLinksHandler(ctx context.Context, req *mcp.CallToolRequest, args BrokenLinksArgs) (*mcp.CallToolResult, any, error) {
	dir := args.Directory

	searchPath := v.GetPath()
	if dir != "" {
		searchPath = filepath.Join(v.GetPath(), dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	existing, err := v.buildExistingNotesSet()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan vault: %v", err)
	}

	var broken []brokenLink
	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		relPath, _ := filepath.Rel(v.GetPath(), path)
		broken = append(broken, findBrokenLinksInNote(relPath, string(content), existing)...)
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan for broken links: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatBrokenLinks(broken)},
		},
	}, nil, nil
}

// noteExists checks if a note exists (handles path normalization)
func (v *Vault) noteExists(link string) bool {
	fullPath := filepath.Join(v.GetPath(), link+".md")
	if _, err := os.Stat(fullPath); err == nil {
		return true
	}

	var found bool
	_ = filepath.Walk(v.GetPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		if name == link {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// normalizeNoteName normalizes a note name for comparison
func normalizeNoteName(link string) string {
	link = strings.TrimSpace(link)
	if idx := strings.Index(link, "#"); idx != -1 {
		link = link[:idx]
	}
	return link
}
