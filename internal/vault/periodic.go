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

// WeeklyNoteHandler gets or creates a weekly note
func (v *Vault) WeeklyNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args WeeklyNoteArgs) (*mcp.CallToolResult, any, error) {
	dateStr := args.Date
	folder := args.Folder
	format := args.Format
	createIfMissing := args.CreateIfMissing

	if folder == "" {
		folder = "weekly"
	}
	if format == "" {
		format = "2006-W02"
	}

	targetDate, err := parseFlexibleDate(dateStr)
	if err != nil {
		return nil, nil, err
	}

	// Get ISO week number
	year, week := targetDate.ISOWeek()
	weekStart := weekStartDate(year, week)

	// Build filename using format or default
	var filename string
	if format == "2006-W02" {
		filename = fmt.Sprintf("%d-W%02d.md", year, week)
	} else {
		filename = weekStart.Format(format) + ".md"
	}

	return v.getOrCreatePeriodicNote(folder, filename, createIfMissing, func() string {
		weekEnd := weekStart.AddDate(0, 0, 6)
		return fmt.Sprintf(`# Week %d, %d

%s - %s

## Goals

- [ ] 

## Notes

## Review

`, week, year, weekStart.Format("Jan 2"), weekEnd.Format("Jan 2, 2006"))
	})
}

// MonthlyNoteHandler gets or creates a monthly note
func (v *Vault) MonthlyNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args MonthlyNoteArgs) (*mcp.CallToolResult, any, error) {
	dateStr := args.Date
	folder := args.Folder
	format := args.Format
	createIfMissing := args.CreateIfMissing

	if folder == "" {
		folder = "monthly"
	}
	if format == "" {
		format = "2006-01"
	}

	targetDate, err := parseFlexibleDate(dateStr)
	if err != nil {
		return nil, nil, err
	}

	// Normalize to first of month
	monthStart := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
	filename := monthStart.Format(format) + ".md"

	return v.getOrCreatePeriodicNote(folder, filename, createIfMissing, func() string {
		return fmt.Sprintf(`# %s

## Goals

- [ ] 

## Weekly Reviews

- [[%d-W%02d]]

## Notes

## Month Review

`, monthStart.Format("January 2006"), monthStart.Year(), getISOWeek(monthStart))
	})
}

// QuarterlyNoteHandler gets or creates a quarterly note
func (v *Vault) QuarterlyNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args QuarterlyNoteArgs) (*mcp.CallToolResult, any, error) {
	dateStr := args.Date
	folder := args.Folder
	createIfMissing := args.CreateIfMissing

	if folder == "" {
		folder = "quarterly"
	}

	targetDate, err := parseFlexibleDate(dateStr)
	if err != nil {
		return nil, nil, err
	}

	quarter := (int(targetDate.Month())-1)/3 + 1
	year := targetDate.Year()
	filename := fmt.Sprintf("%d-Q%d.md", year, quarter)

	return v.getOrCreatePeriodicNote(folder, filename, createIfMissing, func() string {
		return fmt.Sprintf(`# Q%d %d

## Goals

- [ ] 

## Monthly Reviews

- [[%d-01]]
- [[%d-02]]
- [[%d-03]]

## Notes

## Quarter Review

`, quarter, year, year, year, year)
	})
}

// YearlyNoteHandler gets or creates a yearly note
func (v *Vault) YearlyNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args YearlyNoteArgs) (*mcp.CallToolResult, any, error) {
	dateStr := args.Date
	folder := args.Folder
	createIfMissing := args.CreateIfMissing

	if folder == "" {
		folder = "yearly"
	}

	targetDate, err := parseFlexibleDate(dateStr)
	if err != nil {
		return nil, nil, err
	}

	year := targetDate.Year()
	filename := fmt.Sprintf("%d.md", year)

	return v.getOrCreatePeriodicNote(folder, filename, createIfMissing, func() string {
		return fmt.Sprintf(`# %d

## Theme

## Yearly Goals

- [ ] 

## Quarterly Reviews

- [[%d-Q1]]
- [[%d-Q2]]
- [[%d-Q3]]
- [[%d-Q4]]

## Year Review

`, year, year, year, year, year)
	})
}

// ListPeriodicNotesHandler lists periodic notes of a given type
func (v *Vault) ListPeriodicNotesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListPeriodicArgs) (*mcp.CallToolResult, any, error) {
	noteType := args.Type
	limit := args.Limit
	customFolder := args.Folder

	if noteType == "" {
		noteType = "weekly"
	}
	if limit <= 0 {
		limit = 20
	}

	// Map type to default folder
	folderMap := map[string]string{
		"daily":     "daily",
		"weekly":    "weekly",
		"monthly":   "monthly",
		"quarterly": "quarterly",
		"yearly":    "yearly",
	}

	folder, ok := folderMap[noteType]
	if !ok {
		return nil, nil, fmt.Errorf("unknown type: %s. Use: daily, weekly, monthly, quarterly, yearly", noteType)
	}

	// Allow override
	if customFolder != "" {
		folder = customFolder
	}

	searchPath := filepath.Join(v.path, folder)

	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No %s notes folder found: %s", noteType, folder)},
			},
		}, nil, nil
	}

	type noteInfo struct {
		name    string
		path    string
		modTime time.Time
	}

	var notes []noteInfo

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(v.path, path)
			notes = append(notes, noteInfo{
				name:    strings.TrimSuffix(filepath.Base(path), ".md"),
				path:    relPath,
				modTime: info.ModTime(),
			})
		}
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list notes: %v", err)
	}

	if len(notes) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No %s notes found in %s", noteType, folder)},
			},
		}, nil, nil
	}

	// Sort by name descending (most recent first for date-based names)
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].name > notes[j].name
	})

	// Apply limit
	if limit > 0 && limit < len(notes) {
		notes = notes[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d %s notes:\n\n", len(notes), noteType)
	for _, n := range notes {
		fmt.Fprintf(&sb, "- [[%s]] (%s)\n", n.name, n.modTime.Format("Jan 2"))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// Helper: get or create a periodic note
func (v *Vault) getOrCreatePeriodicNote(folder, filename string, create bool, templateFn func() string) (*mcp.CallToolResult, any, error) {
	var notePath string
	if folder != "" {
		notePath = filepath.Join(folder, filename)
	} else {
		notePath = filename
	}

	fullPath := filepath.Join(v.path, notePath)

	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	// Check if exists
	content, err := os.ReadFile(fullPath)
	if err == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Path: %s\n\n%s", notePath, string(content))},
			},
		}, nil, nil
	}

	if !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}

	if !create {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Note not found: %s", notePath)},
			},
		}, nil, nil
	}

	// Create it
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	template := templateFn()
	if err := os.WriteFile(fullPath, []byte(template), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to create note: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Created: %s\n\n%s", notePath, template)},
		},
	}, nil, nil
}

// Helper: parse flexible date input
func parseFlexibleDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}

	formats := []string{
		"2006-01-02",
		"01-02-2006",
		"01/02/2006",
		"2006/01/02",
		"Jan 2, 2006",
		"January 2, 2006",
		"2006-01", // Month only
		"2006",    // Year only
	}

	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

// Helper: get start of ISO week
func weekStartDate(year, week int) time.Time {
	// Jan 4 is always in week 1
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	_, jan4Week := jan4.ISOWeek()

	// Find Monday of week 1
	daysToMonday := int(jan4.Weekday()) - 1
	if jan4.Weekday() == time.Sunday {
		daysToMonday = 6
	}
	week1Monday := jan4.AddDate(0, 0, -daysToMonday)

	// Add weeks
	return week1Monday.AddDate(0, 0, (week-jan4Week)*7)
}

// Helper: get ISO week number
func getISOWeek(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}
