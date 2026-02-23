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

// ListFoldersHandler lists all folders in the vault
func (v *Vault) ListFoldersHandler(ctx context.Context, req *mcp.CallToolRequest, args ListDirsArgs) (*mcp.CallToolResult, any, error) {
	dir := args.Directory
	includeEmpty := args.IncludeEmpty

	searchPath := v.GetPath()
	if dir != "" {
		searchPath = filepath.Join(v.GetPath(), dir)
	}
	if !v.isPathSafe(searchPath) {
		return nil, nil, fmt.Errorf("search path must be within vault")
	}

	type folderInfo struct {
		path      string
		noteCount int
	}

	folders := make(map[string]*folderInfo)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(v.GetPath(), path)
		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			// Skip hidden folders
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			folders[relPath] = &folderInfo{path: relPath}
		} else if strings.HasSuffix(path, ".md") {
			// Count notes in parent folder
			parentDir := filepath.Dir(relPath)
			if parentDir != "." {
				if f, exists := folders[parentDir]; exists {
					f.noteCount++
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list folders: %v", err)
	}

	// Filter and sort
	var result []folderInfo
	for _, f := range folders {
		if includeEmpty || f.noteCount > 0 {
			result = append(result, *f)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].path < result[j].path
	})

	if len(result) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No folders found"},
			},
		}, nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d folders:\n\n", len(result))
	for _, f := range result {
		if f.noteCount > 0 {
			fmt.Fprintf(&sb, "- 📁 %s (%d notes)\n", f.path, f.noteCount)
		} else {
			fmt.Fprintf(&sb, "- 📁 %s (empty)\n", f.path)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// CreateFolderHandler creates a new folder
func (v *Vault) CreateFolderHandler(ctx context.Context, req *mcp.CallToolRequest, args CreateDirArgs) (*mcp.CallToolResult, any, error) {
	folderPath := args.Path

	fullPath := filepath.Join(v.GetPath(), folderPath)

	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	// Check if already exists
	if info, err := os.Stat(fullPath); err == nil {
		if info.IsDir() {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Folder already exists: %s", folderPath)},
				},
			}, nil, nil
		}
		return nil, nil, fmt.Errorf("path exists but is not a folder: %s", folderPath)
	}

	if err := os.MkdirAll(fullPath, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create folder: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Created folder: %s", folderPath)},
		},
	}, nil, nil
}

// MoveNoteHandler moves a note to a new location
func (v *Vault) MoveNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args MoveArgs) (*mcp.CallToolResult, any, error) {
	sourcePath := args.Source
	destPath := args.Destination
	updateLinks := args.UpdateLinks
	dryRun := args.DryRun

	// Ensure .md extension
	if !strings.HasSuffix(sourcePath, ".md") {
		sourcePath += ".md"
	}
	if !strings.HasSuffix(destPath, ".md") {
		destPath += ".md"
	}

	sourceFullPath := filepath.Join(v.GetPath(), sourcePath)
	destFullPath := filepath.Join(v.GetPath(), destPath)

	if !v.isPathSafe(sourceFullPath) || !v.isPathSafe(destFullPath) {
		return nil, nil, fmt.Errorf("paths must be within vault")
	}

	// Check source exists
	if _, err := os.Stat(sourceFullPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("source not found: %s", sourcePath)
	}

	// Check destination doesn't exist
	if _, err := os.Stat(destFullPath); err == nil {
		return nil, nil, fmt.Errorf("destination already exists: %s", destPath)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(destFullPath)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	var updatedFiles int
	if updateLinks && !dryRun {
		updatedFiles = v.updateLinksForMove(sourcePath, destPath)
	}

	if !dryRun {
		// Move the file
		if err := os.Rename(sourceFullPath, destFullPath); err != nil {
			return nil, nil, fmt.Errorf("failed to move note: %v", err)
		}
	}

	var result string
	if updateLinks {
		if dryRun {
			result = fmt.Sprintf("Dry run: would move %s → %s\nWould update links in %d files", sourcePath, destPath, updatedFiles)
		} else {
			result = fmt.Sprintf("Moved %s → %s\nUpdated links in %d files", sourcePath, destPath, updatedFiles)
		}
	} else {
		if dryRun {
			result = fmt.Sprintf("Dry run: would move %s → %s", sourcePath, destPath)
		} else {
			result = fmt.Sprintf("Moved %s → %s", sourcePath, destPath)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

// updateLinksForMove updates wikilinks when a note is moved
func (v *Vault) updateLinksForMove(oldPath, newPath string) int {
	oldName := strings.TrimSuffix(oldPath, ".md")
	oldBase := strings.TrimSuffix(filepath.Base(oldPath), ".md")
	newName := strings.TrimSuffix(newPath, ".md")
	newBase := strings.TrimSuffix(filepath.Base(newPath), ".md")

	updatedFiles := 0

	_ = filepath.Walk(v.GetPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		newContent := contentStr

		// Replace full path links
		newContent = strings.ReplaceAll(newContent, "[["+oldName+"]]", "[["+newName+"]]")
		newContent = strings.ReplaceAll(newContent, "[["+oldName+"|", "[["+newName+"|")

		// Replace basename links (if name changed)
		if oldBase != newBase {
			newContent = strings.ReplaceAll(newContent, "[["+oldBase+"]]", "[["+newBase+"]]")
			newContent = strings.ReplaceAll(newContent, "[["+oldBase+"|", "[["+newBase+"|")
		}

		if newContent != contentStr {
			if err := os.WriteFile(path, []byte(newContent), 0o600); err != nil {
				return nil
			}
			updatedFiles++
		}

		return nil
	})

	return updatedFiles
}

// DeleteFolderHandler deletes an empty folder
func (v *Vault) DeleteFolderHandler(ctx context.Context, req *mcp.CallToolRequest, args DeleteDirArgs) (*mcp.CallToolResult, any, error) {
	folderPath := args.Path
	force := args.Force
	dryRun := args.DryRun

	fullPath := filepath.Join(v.GetPath(), folderPath)

	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("folder not found: %s", folderPath)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("not a folder: %s", folderPath)
	}

	// Check if empty
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read folder: %v", err)
	}

	if len(entries) > 0 && !force {
		return nil, nil, fmt.Errorf("folder not empty: %s (use force=true to delete anyway)", folderPath)
	}

	if dryRun {
		action := "delete folder"
		if force {
			action = "delete folder and contents"
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Dry run: would %s: %s", action, folderPath)},
			},
		}, nil, nil
	}

	if force {
		if err := os.RemoveAll(fullPath); err != nil {
			return nil, nil, fmt.Errorf("failed to delete folder: %v", err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Deleted folder and contents: %s", folderPath)},
			},
		}, nil, nil
	}

	if err := os.Remove(fullPath); err != nil {
		return nil, nil, fmt.Errorf("failed to delete folder: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Deleted folder: %s", folderPath)},
		},
	}, nil, nil
}
