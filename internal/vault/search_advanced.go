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

	"github.com/mark3labs/mcp-go/mcp"
)

// searchMatch represents a search result
type searchMatch struct {
	path    string
	matches []string
	modTime time.Time
}

// getSearchText extracts text to search based on search mode
func getSearchText(content, searchIn string) string {
	switch searchIn {
	case "frontmatter":
		fm := ParseFrontmatter(content)
		var parts []string
		for k, v := range fm {
			parts = append(parts, k+": "+v)
		}
		return strings.Join(parts, "\n")
	case "tags":
		return strings.Join(ExtractTags(content), " ")
	case "all":
		fm := ParseFrontmatter(content)
		var parts []string
		for k, v := range fm {
			parts = append(parts, k+": "+v)
		}
		parts = append(parts, RemoveFrontmatter(content))
		parts = append(parts, ExtractTags(content)...)
		return strings.Join(parts, "\n")
	default: // content
		return RemoveFrontmatter(content)
	}
}

// matchTerms checks which terms match in the search text
func matchTerms(searchText string, terms []string) []string {
	searchTextLower := strings.ToLower(searchText)
	var matched []string
	for _, term := range terms {
		if strings.Contains(searchTextLower, strings.ToLower(term)) {
			matched = append(matched, term)
		}
	}
	return matched
}

// SearchAdvancedHandler performs advanced search with multiple terms
func (v *Vault) SearchAdvancedHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil
	}

	searchIn := req.GetString("in", "content")
	operator := req.GetString("operator", "and")
	dir := req.GetString("directory", "")
	limit := int(req.GetInt("limit", 50))

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	terms := parseSearchTerms(query)
	if len(terms) == 0 {
		return mcp.NewToolResultError("no valid search terms"), nil
	}

	var results []searchMatch

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		searchText := getSearchText(string(content), searchIn)
		matchedTerms := matchTerms(searchText, terms)

		matched := (operator == "or" && len(matchedTerms) > 0) ||
			(operator != "or" && len(matchedTerms) == len(terms))

		if matched {
			relPath, _ := filepath.Rel(v.path, path)
			results = append(results, searchMatch{
				path:    relPath,
				matches: matchedTerms,
				modTime: info.ModTime(),
			})
		}
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No results for: %s", query)), nil
	}

	sort.Slice(results, func(i, j int) bool {
		if len(results[i].matches) != len(results[j].matches) {
			return len(results[i].matches) > len(results[j].matches)
		}
		return results[i].modTime.After(results[j].modTime)
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Search: %s\n\n", query)
	fmt.Fprintf(&sb, "Mode: %s %s | Found %d results\n\n", searchIn, operator, len(results))

	for _, r := range results {
		fmt.Fprintf(&sb, "- **%s** - matched: %s\n", r.path, strings.Join(r.matches, ", "))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// dateResult holds a note path and its date for sorting
type dateResult struct {
	path string
	date time.Time
}

// parseDate attempts to parse a date string using common formats
func parseDate(dateStr string) (time.Time, error) {
	dateFormats := []string{"2006-01-02", "2006/01/02", "Jan 2, 2006", "01-02-2006"}
	for _, format := range dateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// formatDateRange creates a human-readable date range string
func formatDateRange(fromStr, toStr string) string {
	switch {
	case fromStr != "" && toStr != "":
		return fmt.Sprintf("%s to %s", fromStr, toStr)
	case fromStr != "":
		return fmt.Sprintf("from %s", fromStr)
	default:
		return fmt.Sprintf("until %s", toStr)
	}
}

// SearchByDateHandler finds notes by creation/modification date
func (v *Vault) SearchByDateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromStr := req.GetString("from", "")
	toStr := req.GetString("to", "")
	dateType := req.GetString("type", "modified")
	dir := req.GetString("directory", "")
	limit := int(req.GetInt("limit", 50))

	if fromStr == "" && toStr == "" {
		return mcp.NewToolResultError("at least one of 'from' or 'to' is required"), nil
	}

	var fromDate, toDate time.Time
	var err error

	if fromStr != "" {
		fromDate, err = parseDate(fromStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid 'from' date: %s", fromStr)), nil
		}
	}

	if toStr != "" {
		toDate, err = parseDate(toStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid 'to' date: %s", toStr)), nil
		}
		toDate = toDate.Add(24*time.Hour - time.Second) // Include the entire day
	} else {
		toDate = time.Now()
	}

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var results []dateResult

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		checkDate := info.ModTime()

		if !fromDate.IsZero() && checkDate.Before(fromDate) {
			return nil
		}
		if checkDate.After(toDate) {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		results = append(results, dateResult{path: relPath, date: checkDate})
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No notes found in date range"), nil
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].date.After(results[j].date)
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Notes %s (%s)\n\n", dateType, formatDateRange(fromStr, toStr))
	fmt.Fprintf(&sb, "Found %d notes:\n\n", len(results))

	for _, r := range results {
		fmt.Fprintf(&sb, "- **%s** - %s\n", r.path, r.date.Format("Jan 2, 2006"))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// regexMatch holds regex search results for a file
type regexMatch struct {
	path    string
	matches []string
	lines   []int
}

// SearchRegexHandler performs regex pattern search
func (v *Vault) SearchRegexHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, err := req.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError("pattern is required"), nil
	}

	dir := req.GetString("directory", "")
	limit := int(req.GetInt("limit", 50))
	caseInsensitive := req.GetBool("case_insensitive", true)

	if caseInsensitive {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid regex: %v", err)), nil
	}

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var results []regexMatch

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		var matchTexts []string
		var matchLines []int

		for i, line := range lines {
			if found := re.FindAllString(line, -1); len(found) > 0 {
				matchTexts = append(matchTexts, found...)
				matchLines = append(matchLines, i+1)
			}
		}

		if len(matchTexts) > 0 {
			relPath, _ := filepath.Rel(v.path, path)
			results = append(results, regexMatch{
				path:    relPath,
				matches: matchTexts,
				lines:   matchLines,
			})
		}
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No matches for pattern: %s", pattern)), nil
	}

	sort.Slice(results, func(i, j int) bool {
		return len(results[i].matches) > len(results[j].matches)
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Regex Search: %s\n\n", pattern)

	totalMatches := 0
	for _, r := range results {
		totalMatches += len(r.matches)
	}
	fmt.Fprintf(&sb, "Found %d matches in %d files:\n\n", totalMatches, len(results))

	for _, r := range results {
		fmt.Fprintf(&sb, "## %s (%d matches)\n", r.path, len(r.matches))
		if len(r.lines) <= 5 {
			fmt.Fprintf(&sb, "Lines: %v\n", r.lines)
		} else {
			fmt.Fprintf(&sb, "Lines: %v... and %d more\n", r.lines[:5], len(r.lines)-5)
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// parseSearchTerms splits a query into individual search terms
// Handles quoted phrases and strips common operators
func parseSearchTerms(query string) []string {
	var terms []string

	quoteRegex := regexp.MustCompile(`"([^"]+)"`)
	matches := quoteRegex.FindAllStringSubmatch(query, -1)
	for _, m := range matches {
		terms = append(terms, m[1])
	}

	remaining := quoteRegex.ReplaceAllString(query, "")

	for _, word := range strings.Fields(remaining) {
		word = strings.ToLower(word)
		if word == "and" || word == "or" || word == "not" {
			continue
		}
		if len(word) > 1 {
			terms = append(terms, word)
		}
	}

	return terms
}
