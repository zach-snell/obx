package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	// Matches template variables: {{variable}} or {{variable:default}}
	templateVarRegex = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)(?::([^}]*))?\}\}`)
)

// ListTemplatesHandler lists available templates
func (v *Vault) ListTemplatesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListTemplatesArgs) (*mcp.CallToolResult, any, error) {
	folder := args.Folder
	if folder == "" {
		folder = "templates"
	}

	searchPath := filepath.Join(v.path, folder)

	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Templates folder not found: %s", folder)},
			},
		}, nil, nil
	}

	var templates []string

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(searchPath, path)
			templates = append(templates, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list templates: %v", err)
	}

	if len(templates) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No templates found in: %s", folder)},
			},
		}, nil, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d templates in %s:\n\n", len(templates), folder))
	for _, t := range templates {
		sb.WriteString(fmt.Sprintf("- %s\n", t))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// GetTemplateHandler reads a template and shows its variables
func (v *Vault) GetTemplateHandler(ctx context.Context, req *mcp.CallToolRequest, args GetTemplateArgs) (*mcp.CallToolResult, any, error) {
	name := args.Name
	folder := args.Folder
	if folder == "" {
		folder = "templates"
	}

	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}

	templatePath := filepath.Join(v.path, folder, name)

	if !v.isPathSafe(templatePath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("template not found: %s", name)
		}
		return nil, nil, fmt.Errorf("failed to read template: %v", err)
	}

	// Extract variables
	matches := templateVarRegex.FindAllStringSubmatch(string(content), -1)
	vars := make(map[string]string)
	for _, match := range matches {
		varName := match[1]
		defaultVal := ""
		if len(match) > 2 {
			defaultVal = match[2]
		}
		vars[varName] = defaultVal
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Template: %s\n\n", name))

	if len(vars) > 0 {
		sb.WriteString("### Variables:\n")
		for varName, defaultVal := range vars {
			if defaultVal != "" {
				sb.WriteString(fmt.Sprintf("- `{{%s}}` (default: %s)\n", varName, defaultVal))
			} else {
				sb.WriteString(fmt.Sprintf("- `{{%s}}`\n", varName))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Content:\n```markdown\n")
	sb.WriteString(string(content))
	sb.WriteString("\n```")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// ApplyTemplateHandler creates a new note from a template
func (v *Vault) ApplyTemplateHandler(ctx context.Context, req *mcp.CallToolRequest, args ApplyTemplateArgs) (*mcp.CallToolResult, any, error) {
	templateName := args.Template
	targetPath := args.Path
	templateFolder := args.TemplateFolder
	if templateFolder == "" {
		templateFolder = "templates"
	}
	varsStr := args.Variables

	if !strings.HasSuffix(templateName, ".md") {
		templateName += ".md"
	}
	if !strings.HasSuffix(targetPath, ".md") {
		targetPath += ".md"
	}

	// Read template
	templatePath := filepath.Join(v.path, templateFolder, templateName)
	if !v.isPathSafe(templatePath) {
		return nil, nil, fmt.Errorf("template path must be within vault")
	}

	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("template not found: %s", templateName)
		}
		return nil, nil, fmt.Errorf("failed to read template: %v", err)
	}

	// Check target doesn't exist
	fullTargetPath := filepath.Join(v.path, targetPath)
	if !v.isPathSafe(fullTargetPath) {
		return nil, nil, fmt.Errorf("target path must be within vault")
	}

	if _, err := os.Stat(fullTargetPath); err == nil {
		return nil, nil, fmt.Errorf("target note already exists: %s", targetPath)
	}

	// Parse variables from string (format: "key1=value1,key2=value2")
	userVars := make(map[string]string)
	if varsStr != "" {
		for _, pair := range strings.Split(varsStr, ",") {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 {
				userVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Add built-in variables
	now := time.Now()
	builtinVars := map[string]string{
		"date":      now.Format("2006-01-02"),
		"time":      now.Format("15:04"),
		"datetime":  now.Format("2006-01-02 15:04"),
		"year":      now.Format("2006"),
		"month":     now.Format("01"),
		"day":       now.Format("02"),
		"title":     strings.TrimSuffix(filepath.Base(targetPath), ".md"),
		"filename":  filepath.Base(targetPath),
		"folder":    filepath.Dir(targetPath),
		"timestamp": fmt.Sprintf("%d", now.Unix()),
	}

	// Merge variables (user vars override builtins)
	for k, v := range builtinVars {
		if _, exists := userVars[k]; !exists {
			userVars[k] = v
		}
	}

	// Apply template substitution
	result := templateVarRegex.ReplaceAllStringFunc(string(templateContent), func(match string) string {
		submatch := templateVarRegex.FindStringSubmatch(match)
		varName := submatch[1]
		defaultVal := ""
		if len(submatch) > 2 {
			defaultVal = submatch[2]
		}

		if val, ok := userVars[varName]; ok {
			return val
		}
		if defaultVal != "" {
			return defaultVal
		}
		return match // Keep original if no value
	})

	// Create target directory if needed
	targetDir := filepath.Dir(fullTargetPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Write the new note
	if err := os.WriteFile(fullTargetPath, []byte(result), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to create note: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Created note from template:\n- Template: %s\n- Target: %s\n\n%s",
				templateName, targetPath, result)},
		},
	}, nil, nil
}
