package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// matchNoteByScope checks whether a note matches the search terms within the given scope.
func matchNoteByScope(searchIn, relPath, contentStr string, terms []string, operator string) (matched bool, matchLine int, matchContent string) {
	switch searchIn {
	case "file":
		if matchTerms(relPath, terms, operator) {
			return true, 0, relPath
		}
	case "heading":
		for _, h := range extractHeadings(contentStr) {
			if matchTerms(h.Text, terms, operator) {
				return true, h.Line, h.Text
			}
		}
	default: // content
		for i, line := range strings.Split(contentStr, "\n") {
			if matchTerms(line, terms, operator) {
				return true, i + 1, strings.TrimSpace(line)
			}
		}
	}
	return false, 0, ""
}

// SearchAdvancedHandler performs advanced search with operators and scope
func (v *Vault) SearchAdvancedHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchAdvancedArgs) (*mcp.CallToolResult, any, error) {
	searchIn := args.SearchIn
	operator := args.Operator
	limit := args.Limit

	if searchIn == "" {
		searchIn = "content"
	}
	if operator == "" {
		operator = "and"
	}
	if limit <= 0 {
		limit = 50
	}

	searchPath := v.path
	if args.Directory != "" {
		searchPath = filepath.Join(v.path, args.Directory)
	}

	terms := parseSearchTerms(args.Query)
	if len(terms) == 0 {
		return nil, nil, fmt.Errorf("empty search query")
	}

	var results []SearchResult

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		if matched, line, text := matchNoteByScope(searchIn, relPath, string(content), terms, operator); matched {
			results = append(results, SearchResult{File: relPath, Line: line, Content: text})
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %v", err)
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No matches found for: %s", args.Query)},
			},
		}, nil, nil
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatSearchResults(results, args.Query)},
		},
	}, nil, nil
}

// formatSearchResults formats search results grouped by file.
func formatSearchResults(results []SearchResult, query string) string {
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
		if r.Line > 0 {
			sb.WriteString(fmt.Sprintf("  L%d: %s\n", r.Line, truncate(r.Content, 100)))
		}
	}
	return sb.String()
}

// parseDateRange parses from/to date strings and returns the time range.
func parseDateRange(fromStr, toStr string) (fromTime, toTime time.Time, err error) {
	if fromStr != "" {
		fromTime, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return fromTime, toTime, fmt.Errorf("invalid from date: %v", err)
		}
	}
	if toStr != "" {
		toTime, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return fromTime, toTime, fmt.Errorf("invalid to date: %v", err)
		}
		toTime = toTime.Add(24*time.Hour - time.Nanosecond)
	}
	return fromTime, toTime, nil
}

// isTimeInRange checks whether t falls within the [from, to] range.
// Zero-value bounds are treated as unbounded.
func isTimeInRange(t, from, to time.Time) bool {
	if !from.IsZero() && t.Before(from) {
		return false
	}
	if !to.IsZero() && t.After(to) {
		return false
	}
	return true
}

type dateResult struct {
	path string
	time time.Time
}

// SearchDateHandler searches notes by date (created or modified)
func (v *Vault) SearchDateHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchDateArgs) (*mcp.CallToolResult, any, error) {
	limit := args.Limit
	if limit <= 0 {
		limit = 50
	}

	fromTime, toTime, err := parseDateRange(args.From, args.To)
	if err != nil {
		return nil, nil, err
	}

	searchPath := v.path
	if args.Directory != "" {
		searchPath = filepath.Join(v.path, args.Directory)
	}

	// Note: "created" date type falls back to modtime since Go's os.FileInfo
	// does not expose creation time cross-platform. A future improvement could
	// read creation dates from frontmatter.
	var results []dateResult

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if !isTimeInRange(info.ModTime(), fromTime, toTime) {
			return nil
		}
		relPath, _ := filepath.Rel(v.path, path)
		results = append(results, dateResult{path: relPath, time: info.ModTime()})
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %v", err)
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No notes found in date range"},
			},
		}, nil, nil
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].time.After(results[j].time)
	})

	if len(results) > limit {
		results = results[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes:\n\n", len(results)))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", r.path, r.time.Format("2006-01-02 15:04")))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// SearchRegexHandler searches using regex
func (v *Vault) SearchRegexHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchRegexArgs) (*mcp.CallToolResult, any, error) {
	pattern := args.Pattern
	dir := args.Directory
	limit := args.Limit
	caseInsensitive := args.CaseInsensitive

	if limit <= 0 {
		limit = 50
	}

	if caseInsensitive && !strings.HasPrefix(pattern, "(?i)") {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid regex: %v", err)
	}

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var results []SearchResult

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		relPath, _ := filepath.Rel(v.path, path)

		for i, line := range lines {
			if re.MatchString(line) {
				results = append(results, SearchResult{
					File:    relPath,
					Line:    i + 1,
					Content: strings.TrimSpace(line),
				})
				// Limit matches per file? No, maybe all matches.
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
				&mcp.TextContent{Text: fmt.Sprintf("No matches found for: %s", pattern)},
			},
		}, nil, nil
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d matches for %s:\n\n", len(results), pattern))

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

// parseSearchTerms splits query into terms, handling quotes
func parseSearchTerms(query string) []string {
	var terms []string
	var current strings.Builder
	inQuote := false

	for _, c := range query {
		switch {
		case c == '"':
			inQuote = !inQuote
			if !inQuote && current.Len() > 0 {
				terms = append(terms, current.String())
				current.Reset()
			}
		case c == ' ' && !inQuote:
			if current.Len() > 0 {
				terms = append(terms, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(c)
		}
	}
	if current.Len() > 0 {
		terms = append(terms, current.String())
	}

	// Filter common stopwords or short terms?
	var filtered []string
	for _, t := range terms {
		t = strings.TrimSpace(t)
		if len(t) > 1 {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func matchTerms(text string, terms []string, operator string) bool {
	textLower := strings.ToLower(text)
	matches := 0
	for _, term := range terms {
		// Handle NOT operator (if term starts with -?)
		// Simple implementation for now
		if strings.Contains(textLower, strings.ToLower(term)) {
			matches++
		}
	}

	if operator == "or" {
		return matches > 0
	}
	// Default AND
	return matches == len(terms)
}
