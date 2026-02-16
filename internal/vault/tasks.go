package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Task represents a parsed task from markdown
type Task struct {
	File      string   `json:"file"`
	Line      int      `json:"line"`
	Completed bool     `json:"completed"`
	Text      string   `json:"text"`
	DueDate   *string  `json:"dueDate,omitempty"`
	Priority  *string  `json:"priority,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

var (
	// Matches: - [ ] or - [x] or - [X]
	taskRegex = regexp.MustCompile(`^(\s*)-\s*\[([ xX])\]\s*(.+)$`)
	// Matches: 📅 2024-01-15
	dueDateRegex = regexp.MustCompile(`📅\s*(\d{4}-\d{2}-\d{2})`)
	// Matches: ⏫ (high), 🔼 (medium), 🔽 (low)
	priorityRegex = regexp.MustCompile(`([⏫🔼🔽])`)
	// Matches: #tag
	tagRegex = regexp.MustCompile(`#([a-zA-Z0-9_\-]+)`)
)

// ParseTask parses a single line into a Task if it matches
func ParseTask(line string, lineNum int) *Task {
	match := taskRegex.FindStringSubmatch(line)
	if match == nil {
		return nil
	}

	status := match[2]
	text := strings.TrimSpace(match[3])

	task := &Task{
		Line:      lineNum,
		Completed: strings.EqualFold(status, "x"),
		Text:      text,
	}

	// Extract due date
	if dateMatch := dueDateRegex.FindStringSubmatch(text); dateMatch != nil {
		task.DueDate = &dateMatch[1]
	}

	// Extract priority
	if prioMatch := priorityRegex.FindStringSubmatch(text); prioMatch != nil {
		var prio string
		switch prioMatch[1] {
		case "⏫":
			prio = "high"
		case "🔼":
			prio = "medium"
		case "🔽":
			prio = "low"
		}
		task.Priority = &prio
	}

	// Extract tags
	tagMatches := tagRegex.FindAllStringSubmatch(text, -1)
	for _, tm := range tagMatches {
		task.Tags = append(task.Tags, tm[1])
	}

	return task
}

// taskMatchesStatus returns whether a task should be included given the status filter.
func taskMatchesStatus(task *Task, status string) bool {
	switch status {
	case "open":
		return !task.Completed
	case "completed":
		return task.Completed
	default: // "all"
		return true
	}
}

// collectTasks walks a directory and collects tasks matching the given status filter.
func (v *Vault) collectTasks(searchPath, status string) ([]Task, error) {
	var tasks []Task

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		relPath, _ := filepath.Rel(v.path, path)
		for i, line := range strings.Split(string(content), "\n") {
			if task := ParseTask(line, i+1); task != nil && taskMatchesStatus(task, status) {
				task.File = relPath
				tasks = append(tasks, *task)
			}
		}
		return nil
	})
	return tasks, err
}

// formatTasks formats a slice of tasks grouped by file.
func formatTasks(tasks []Task) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d tasks:\n\n", len(tasks)))

	currentFile := ""
	for _, t := range tasks {
		if t.File != currentFile {
			if currentFile != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("## %s\n", t.File))
			currentFile = t.File
		}

		checkbox := "[ ]"
		if t.Completed {
			checkbox = "[x]"
		}

		sb.WriteString(fmt.Sprintf("  L%d: - %s %s", t.Line, checkbox, t.Text))
		if t.Priority != nil {
			sb.WriteString(fmt.Sprintf(" [%s]", *t.Priority))
		}
		if t.DueDate != nil {
			sb.WriteString(fmt.Sprintf(" (due: %s)", *t.DueDate))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// ListTasksHandler lists all tasks across the vault
func (v *Vault) ListTasksHandler(ctx context.Context, req *mcp.CallToolRequest, args ListTasksArgs) (*mcp.CallToolResult, any, error) {
	status := args.Status
	if status == "" {
		status = "all"
	}

	searchPath := v.path
	if args.Directory != "" {
		searchPath = filepath.Join(v.path, args.Directory)
	}

	tasks, err := v.collectTasks(searchPath, status)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tasks: %v", err)
	}

	if len(tasks) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No tasks found"},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatTasks(tasks)},
		},
	}, nil, nil
}

// ToggleTaskHandler toggles a task's completion status
func (v *Vault) ToggleTaskHandler(ctx context.Context, req *mcp.CallToolRequest, args ToggleTaskArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	line := args.Line

	if !strings.HasSuffix(path, ".md") {
		return nil, nil, fmt.Errorf("path must end with .md")
	}

	fullPath := filepath.Join(v.path, path)

	// Security: ensure path is within vault
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("note not found: %s", path)
		}
		return nil, nil, fmt.Errorf("failed to read note: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	if line < 1 || line > len(lines) {
		return nil, nil, fmt.Errorf("line %d out of range (1-%d)", line, len(lines))
	}

	targetLine := lines[line-1]
	task := ParseTask(targetLine, line)
	if task == nil {
		return nil, nil, fmt.Errorf("line %d is not a task", line)
	}

	// Toggle the checkbox
	var newLine string
	if task.Completed {
		// [x] or [X] -> [ ]
		newLine = strings.Replace(targetLine, "[x]", "[ ]", 1)
		newLine = strings.Replace(newLine, "[X]", "[ ]", 1)
	} else {
		// [ ] -> [x]
		newLine = strings.Replace(targetLine, "[ ]", "[x]", 1)
	}

	lines[line-1] = newLine

	if err := os.WriteFile(fullPath, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	newStatus := "completed"
	if task.Completed {
		newStatus = "open"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Toggled task on line %d to %s: %s", line, newStatus, task.Text)},
		},
	}, nil, nil
}
