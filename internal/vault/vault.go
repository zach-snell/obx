package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Vault represents an Obsidian vault
type Vault struct {
	mu            sync.RWMutex
	activePath    string
	allowedVaults map[string]string
}

// New creates a new Vault instance
func New(path string) *Vault {
	cleanPath, _ := filepath.Abs(filepath.Clean(path))
	return &Vault{
		activePath:    cleanPath,
		allowedVaults: make(map[string]string),
	}
}

// SetAllowedVaults configures which vaults can be dynamically switched to
func (v *Vault) SetAllowedVaults(vaults map[string]string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.allowedVaults = vaults
}

// GetAllowedVaults thread-safely returns the configured vaults
func (v *Vault) GetAllowedVaults() map[string]string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Return a copy to prevent external mutation
	vaults := make(map[string]string, len(v.allowedVaults))
	for k, val := range v.allowedVaults {
		vaults[k] = val
	}
	return vaults
}

// GetPath thread-safely returns the active vault path
func (v *Vault) GetPath() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.activePath
}

// SetPath thread-safely sets the active vault path
func (v *Vault) SetPath(path string) {
	cleanPath, _ := filepath.Abs(filepath.Clean(path))
	v.mu.Lock()
	defer v.mu.Unlock()
	v.activePath = cleanPath
}

// isPathSafe checks if the given path is within the vault (prevents path traversal)
func (v *Vault) isPathSafe(fullPath string) bool {
	cleanVault := filepath.Clean(v.GetPath())
	cleanTarget := filepath.Clean(fullPath)

	// First apply lexical path traversal protection.
	if !isPathWithinBase(cleanVault, cleanTarget) {
		return false
	}

	// Then apply symlink-aware protection to catch vault-internal symlink escapes.
	resolvedVault := resolvePathWithExistingAncestors(cleanVault)
	resolvedTarget := resolvePathWithExistingAncestors(cleanTarget)
	return isPathWithinBase(resolvedVault, resolvedTarget)
}

func isPathWithinBase(basePath, targetPath string) bool {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func resolvePathWithExistingAncestors(targetPath string) string {
	existingPath := targetPath
	var tail []string

	for {
		if _, err := os.Lstat(existingPath); err == nil {
			break
		}

		parent := filepath.Dir(existingPath)
		if parent == existingPath {
			break
		}

		tail = append([]string{filepath.Base(existingPath)}, tail...)
		existingPath = parent
	}

	resolvedPath := existingPath
	if resolved, err := filepath.EvalSymlinks(existingPath); err == nil {
		resolvedPath = resolved
	}

	for _, segment := range tail {
		resolvedPath = filepath.Join(resolvedPath, segment)
	}

	return filepath.Clean(resolvedPath)
}

// ListNotesHandler lists all notes in the vault with optional pagination
func (v *Vault) ListNotesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListNotesArgs) (*mcp.CallToolResult, any, error) {
	dir := args.Directory
	limit := args.Limit
	offset := args.Offset
	mode := normalizeMode(args.Mode)
	if !isDetailedMode(mode) && limit <= 0 {
		// Keep compact mode bounded by default.
		limit = 100
	}

	searchPath := v.GetPath()
	if dir != "" {
		searchPath = filepath.Join(v.GetPath(), dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	var notes []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(v.GetPath(), path)
			notes = append(notes, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list notes: %v", err)
	}

	totalCount := len(notes)
	if totalCount == 0 {
		if !isDetailedMode(mode) {
			return compactResult("No notes found", false, map[string]any{
				"total_count":    0,
				"returned_count": 0,
				"offset":         offset,
				"limit":          limit,
				"notes":          []string{},
			}, nil)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No notes found"},
			},
		}, nil, nil
	}

	// Apply pagination
	if offset > 0 {
		if offset >= totalCount {
			if !isDetailedMode(mode) {
				return compactResult(
					fmt.Sprintf("Offset %d exceeds total count %d", offset, totalCount),
					false,
					map[string]any{
						"total_count":    totalCount,
						"returned_count": 0,
						"offset":         offset,
						"limit":          limit,
						"notes":          []string{},
					},
					nil,
				)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Offset %d exceeds total count %d", offset, totalCount)},
				},
			}, nil, nil
		}
		notes = notes[offset:]
	}

	truncated := false
	if limit > 0 && limit < len(notes) {
		notes = notes[:limit]
		truncated = true
	}

	if !isDetailedMode(mode) {
		next := map[string]any(nil)
		if offset+len(notes) < totalCount {
			next = map[string]any{
				"tool": "manage-notes",
				"args": map[string]any{
					"action":    "list",
					"directory": dir,
					"offset":    offset + len(notes),
					"limit":     limit,
					"mode":      modeCompact,
				},
			}
		}

		return compactResult(
			fmt.Sprintf("Found %d notes", totalCount),
			truncated,
			map[string]any{
				"total_count":    totalCount,
				"returned_count": len(notes),
				"offset":         offset,
				"limit":          limit,
				"notes":          notes,
			},
			next,
		)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes", totalCount))
	if offset > 0 || limit > 0 {
		sb.WriteString(fmt.Sprintf(" (showing %d-%d)", offset+1, offset+len(notes)))
	}
	sb.WriteString(":\n\n")
	sb.WriteString(strings.Join(notes, "\n"))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// ReadNoteHandler reads a note's content
func (v *Vault) ReadNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args ReadNoteArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	if !strings.HasSuffix(path, ".md") {
		return nil, nil, fmt.Errorf("path must end with .md")
	}

	fullPath := filepath.Join(v.GetPath(), path)

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

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(content)},
		},
	}, nil, nil
}

// WriteNoteHandler creates or updates a note
func (v *Vault) WriteNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args WriteNoteArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	content := args.Content
	expectedMtime := args.ExpectedMtime

	if !strings.HasSuffix(path, ".md") {
		return nil, nil, fmt.Errorf("path must end with .md")
	}

	fullPath := filepath.Join(v.GetPath(), path)

	// Security: ensure path is within vault
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Optimistic concurrency check for existing files.
	if _, err := os.Stat(fullPath); err == nil {
		if err := ensureExpectedMtime(fullPath, expectedMtime); err != nil {
			return nil, nil, err
		}
	} else if expectedMtime != "" && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("failed to stat note: %v", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Successfully wrote: %s", path)},
		},
	}, nil, nil
}

// DeleteNoteHandler deletes a note
func (v *Vault) DeleteNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args DeleteNoteArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	dryRun := args.DryRun
	expectedMtime := args.ExpectedMtime
	if !strings.HasSuffix(path, ".md") {
		return nil, nil, fmt.Errorf("path must end with .md")
	}

	fullPath := filepath.Join(v.GetPath(), path)

	// Security: ensure path is within vault
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	if err := ensureExpectedMtime(fullPath, expectedMtime); err != nil {
		return nil, nil, err
	}

	if dryRun {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Dry run: would delete %s", path)},
			},
		}, nil, nil
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("note not found: %s", path)
		}
		return nil, nil, fmt.Errorf("failed to delete note: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Successfully deleted: %s", path)},
		},
	}, nil, nil
}
