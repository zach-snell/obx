package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DailyNoteHandler gets or creates a daily note
func (v *Vault) DailyNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args DailyNoteArgs) (*mcp.CallToolResult, any, error) {
	dateStr := args.Date
	folder := args.Folder
	format := args.Format
	createIfMissing := args.CreateIfMissing

	if folder == "" {
		folder = "daily"
	}
	if format == "" {
		format = "2006-01-02"
	}

	targetDate, err := parseFlexibleDate(dateStr)
	if err != nil {
		return nil, nil, err
	}

	filename := targetDate.Format(format) + ".md"

	return v.getOrCreatePeriodicNote(folder, filename, createIfMissing, func() string {
		return fmt.Sprintf(`# %s

## Goals

- [ ] 

## Notes

## Review

`, targetDate.Format("Monday, January 2, 2006"))
	})
}

// ListDailyNotesHandler lists daily notes in a date range
func (v *Vault) ListDailyNotesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListPeriodicArgs) (*mcp.CallToolResult, any, error) {
	// Reusing ListPeriodicArgs which has Limit and Folder
	folder := args.Folder
	limit := args.Limit

	if folder == "" {
		folder = "daily"
	}
	if limit <= 0 {
		limit = 30
	}

	searchPath := v.path
	if folder != "" {
		searchPath = filepath.Join(v.path, folder)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	type noteInfo struct {
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
			notes = append(notes, noteInfo{path: relPath, modTime: info.ModTime()})
		}
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list notes: %v", err)
	}

	if len(notes) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No daily notes found"},
			},
		}, nil, nil
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(notes)-1; i++ {
		for j := i + 1; j < len(notes); j++ {
			if notes[j].modTime.After(notes[i].modTime) {
				notes[i], notes[j] = notes[j], notes[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && limit < len(notes) {
		notes = notes[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d daily notes:\n\n", len(notes)))
	for _, n := range notes {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", n.path, n.modTime.Format("Jan 2, 2006")))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}
