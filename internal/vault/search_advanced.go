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

// SearchAdvancedHandler performs advanced search with operators and scope
func (v *Vault) SearchAdvancedHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchAdvancedArgs) (*mcp.CallToolResult, any, error) {
	query := args.Query
	searchIn := args.SearchIn
	operator := args.Operator
	dir := args.Directory
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
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	// Parse query terms
	terms := parseSearchTerms(query)
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

		contentStr := string(content)
		relPath, _ := filepath.Rel(v.path, path)

		// Check match based on scope
		matched := false
		matchLine := 0
		matchContent := ""

		switch searchIn {
		case "file":
			if matchTerms(relPath, terms, operator) {
				matched = true
				matchContent = relPath
			}
		case "heading":
			headings := extractHeadings(contentStr) // helper from batch.go
			for _, h := range headings {
				if matchTerms(h.Text, terms, operator) {
					matched = true
					matchLine = h.Line
					matchContent = h.Text
					break
				}
			}
		default: // content
			lines := strings.Split(contentStr, "\n")
			for i, line := range lines {
				if matchTerms(line, terms, operator) {
					matched = true
					matchLine = i + 1
					matchContent = strings.TrimSpace(line)
					break // Just find first match per file for now
				}
			}
		}

		if matched {
			results = append(results, SearchResult{
				File:    relPath,
				Line:    matchLine,
				Content: matchContent,
			})
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %v", err)
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No matches found for: %s", query)},
			},
		}, nil, nil
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

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

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// SearchDateHandler searches notes by date (created or modified)
func (v *Vault) SearchDateHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchDateArgs) (*mcp.CallToolResult, any, error) {
	fromStr := args.From
	toStr := args.To
	dateType := args.DateType
	dir := args.Directory
	limit := args.Limit

	if dateType == "" {
		dateType = "modified"
	}
	if limit <= 0 {
		limit = 50
	}

	var fromTime, toTime time.Time
	var err error

	if fromStr != "" {
		fromTime, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid from date: %v", err)
		}
	}
	if toStr != "" {
		toTime, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid to date: %v", err)
		}
		// Set to end of day
		toTime = toTime.Add(24*time.Hour - time.Nanosecond)
	}

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	type dateResult struct {
		path string
		time time.Time
	}
	var results []dateResult

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		checkTime := info.ModTime()
		// Go doesn't provide creation time cross-platform easily in os.FileInfo
		// So we fallback to modtime or skip creation time check strictly
		if dateType == "created" {
			// In a real implementation, we might use syscall or frontmatter
			// For now, warn or fallback
		}

		if !fromTime.IsZero() && checkTime.Before(fromTime) {
			return nil
		}
		if !toTime.IsZero() && checkTime.After(toTime) {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		results = append(results, dateResult{path: relPath, time: checkTime})
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

	// Sort by date descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].time.After(results[j].time)
	})

	if limit > 0 && len(results) > limit {
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
		if c == '"' {
			inQuote = !inQuote
			if !inQuote && current.Len() > 0 {
				terms = append(terms, current.String())
				current.Reset()
			}
		} else if c == ' ' && !inQuote {
			if current.Len() > 0 {
				terms = append(terms, current.String())
				current.Reset()
			}
		} else {
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
