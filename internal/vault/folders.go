package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListFoldersHandler lists all folders in the vault
func (v *Vault) ListFoldersHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir := req.GetString("directory", "")
	includeEmpty := req.GetBool("include_empty", true)

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
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

		relPath, _ := filepath.Rel(v.path, path)
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
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list folders: %v", err)), nil
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
		return mcp.NewToolResultText("No folders found"), nil
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

	return mcp.NewToolResultText(sb.String()), nil
}

// CreateFolderHandler creates a new folder
func (v *Vault) CreateFolderHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	folderPath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	fullPath := filepath.Join(v.path, folderPath)

	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Check if already exists
	if info, err := os.Stat(fullPath); err == nil {
		if info.IsDir() {
			return mcp.NewToolResultText(fmt.Sprintf("Folder already exists: %s", folderPath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Path exists but is not a folder: %s", folderPath)), nil
	}

	if err := os.MkdirAll(fullPath, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create folder: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Created folder: %s", folderPath)), nil
}

// MoveNoteHandler moves a note to a new location
func (v *Vault) MoveNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, err := req.RequireString("source")
	if err != nil {
		return mcp.NewToolResultError("source path is required"), nil
	}

	destPath, err := req.RequireString("destination")
	if err != nil {
		return mcp.NewToolResultError("destination path is required"), nil
	}

	updateLinks := req.GetBool("update_links", true)

	// Ensure .md extension
	if !strings.HasSuffix(sourcePath, ".md") {
		sourcePath += ".md"
	}
	if !strings.HasSuffix(destPath, ".md") {
		destPath += ".md"
	}

	sourceFullPath := filepath.Join(v.path, sourcePath)
	destFullPath := filepath.Join(v.path, destPath)

	if !v.isPathSafe(sourceFullPath) || !v.isPathSafe(destFullPath) {
		return mcp.NewToolResultError("paths must be within vault"), nil
	}

	// Check source exists
	if _, err := os.Stat(sourceFullPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Source not found: %s", sourcePath)), nil
	}

	// Check destination doesn't exist
	if _, err := os.Stat(destFullPath); err == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Destination already exists: %s", destPath)), nil
	}

	// Create destination directory if needed
	destDir := filepath.Dir(destFullPath)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	var updatedFiles int
	if updateLinks {
		updatedFiles = v.updateLinksForMove(sourcePath, destPath)
	}

	// Move the file
	if err := os.Rename(sourceFullPath, destFullPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to move note: %v", err)), nil
	}

	var result string
	if updateLinks {
		result = fmt.Sprintf("Moved %s → %s\nUpdated links in %d files", sourcePath, destPath, updatedFiles)
	} else {
		result = fmt.Sprintf("Moved %s → %s", sourcePath, destPath)
	}

	return mcp.NewToolResultText(result), nil
}

// updateLinksForMove updates wikilinks when a note is moved
func (v *Vault) updateLinksForMove(oldPath, newPath string) int {
	oldName := strings.TrimSuffix(oldPath, ".md")
	oldBase := strings.TrimSuffix(filepath.Base(oldPath), ".md")
	newName := strings.TrimSuffix(newPath, ".md")
	newBase := strings.TrimSuffix(filepath.Base(newPath), ".md")

	updatedFiles := 0

	_ = filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
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
func (v *Vault) DeleteFolderHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	folderPath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	force := req.GetBool("force", false)

	fullPath := filepath.Join(v.path, folderPath)

	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Folder not found: %s", folderPath)), nil
	}
	if !info.IsDir() {
		return mcp.NewToolResultError(fmt.Sprintf("Not a folder: %s", folderPath)), nil
	}

	// Check if empty
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read folder: %v", err)), nil
	}

	if len(entries) > 0 && !force {
		return mcp.NewToolResultError(fmt.Sprintf("Folder not empty: %s (use force=true to delete anyway)", folderPath)), nil
	}

	if force {
		if err := os.RemoveAll(fullPath); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete folder: %v", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Deleted folder and contents: %s", folderPath)), nil
	}

	if err := os.Remove(fullPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete folder: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Deleted folder: %s", folderPath)), nil
}
