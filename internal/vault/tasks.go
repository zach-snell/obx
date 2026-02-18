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
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
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

// findTaskByText finds a task in lines by partial text match.
// Returns the 1-based line number and the parsed task, or an error.
func findTaskByText(lines []string, text string) (int, *Task, error) {
	textLower := strings.ToLower(strings.TrimSpace(text))
	var matches []int

	for i, line := range lines {
		task := ParseTask(line, i+1)
		if task != nil && strings.Contains(strings.ToLower(task.Text), textLower) {
			matches = append(matches, i+1)
		}
	}

	switch len(matches) {
	case 0:
		return 0, nil, fmt.Errorf("no task matching %q found", text)
	case 1:
		lineNum := matches[0]
		task := ParseTask(lines[lineNum-1], lineNum)
		return lineNum, task, nil
	default:
		return 0, nil, fmt.Errorf("ambiguous: %d tasks match %q — provide more specific text", len(matches), text)
	}
}

// toggleLine toggles a task checkbox on a single line and returns the new line.
func toggleLine(line string, task *Task) string {
	if task.Completed {
		newLine := strings.Replace(line, "[x]", "[ ]", 1)
		return strings.Replace(newLine, "[X]", "[ ]", 1)
	}
	return strings.Replace(line, "[ ]", "[x]", 1)
}

// ToggleTaskHandler toggles a task's completion status by line number or text match.
func (v *Vault) ToggleTaskHandler(ctx context.Context, req *mcp.CallToolRequest, args ToggleTaskArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	if !strings.HasSuffix(path, ".md") {
		path += ".md"
	}

	fullPath := filepath.Join(v.path, path)
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

	// Resolve the target task — by text or line number.
	var lineNum int
	var task *Task

	switch {
	case args.Text != "":
		lineNum, task, err = findTaskByText(lines, args.Text)
		if err != nil {
			return nil, nil, err
		}
	case args.Line > 0:
		lineNum = args.Line
		if lineNum < 1 || lineNum > len(lines) {
			return nil, nil, fmt.Errorf("line %d out of range (1-%d)", lineNum, len(lines))
		}
		task = ParseTask(lines[lineNum-1], lineNum)
		if task == nil {
			return nil, nil, fmt.Errorf("line %d is not a task", lineNum)
		}
	default:
		return nil, nil, fmt.Errorf("either 'line' or 'text' must be provided")
	}

	lines[lineNum-1] = toggleLine(lines[lineNum-1], task)

	if err := os.WriteFile(fullPath, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	newStatus := "completed"
	if task.Completed {
		newStatus = "open"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Toggled task on L%d to %s: %s", lineNum, newStatus, task.Text)},
		},
	}, nil, nil
}

// CompleteTasksHandler marks multiple tasks as complete by text match.
func (v *Vault) CompleteTasksHandler(ctx context.Context, req *mcp.CallToolRequest, args CompleteTasksArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	if !strings.HasSuffix(path, ".md") {
		path += ".md"
	}

	fullPath := filepath.Join(v.path, path)
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

	texts := parsePaths(args.Texts) // reuse comma/JSON array parser
	if len(texts) == 0 {
		return nil, nil, fmt.Errorf("texts is required")
	}

	lines := strings.Split(string(content), "\n")

	var completed []string
	var errors []string

	for _, text := range texts {
		lineNum, task, err := findTaskByText(lines, strings.TrimSpace(text))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%q: %v", text, err))
			continue
		}
		if task.Completed {
			completed = append(completed, fmt.Sprintf("L%d: %s (already complete)", lineNum, task.Text))
			continue
		}
		lines[lineNum-1] = strings.Replace(lines[lineNum-1], "[ ]", "[x]", 1)
		completed = append(completed, fmt.Sprintf("L%d: %s", lineNum, task.Text))
	}

	if err := os.WriteFile(fullPath, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Completed %d task(s) in %s:\n", len(completed), path)
	for _, c := range completed {
		fmt.Fprintf(&sb, "  ✓ %s\n", c)
	}
	if len(errors) > 0 {
		sb.WriteString("\nCould not match:\n")
		for _, e := range errors {
			fmt.Fprintf(&sb, "  ✗ %s\n", e)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}
