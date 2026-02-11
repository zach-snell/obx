package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// Canvas represents an Obsidian canvas file
type Canvas struct {
	Nodes []CanvasNode `json:"nodes"`
	Edges []CanvasEdge `json:"edges"`
}

// CanvasNode represents a node in a canvas
type CanvasNode struct {
	ID     string `json:"id"`
	Type   string `json:"type"` // text, file, link, group
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`

	// For text nodes
	Text string `json:"text,omitempty"`

	// For file nodes
	File string `json:"file,omitempty"`

	// For link nodes
	URL string `json:"url,omitempty"`

	// For group nodes
	Label string `json:"label,omitempty"`

	// Optional styling
	Color string `json:"color,omitempty"`
}

// CanvasEdge represents a connection between nodes
type CanvasEdge struct {
	ID       string `json:"id"`
	FromNode string `json:"fromNode"`
	ToNode   string `json:"toNode"`
	FromSide string `json:"fromSide,omitempty"` // top, right, bottom, left
	ToSide   string `json:"toSide,omitempty"`
	Label    string `json:"label,omitempty"`
}

// ListCanvasesHandler lists all canvas files in the vault
func (v *Vault) ListCanvasesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir := req.GetString("directory", "")

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var canvases []string

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".canvas") {
			relPath, _ := filepath.Rel(v.path, path)
			canvases = append(canvases, relPath)
		}
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list canvases: %v", err)), nil
	}

	if len(canvases) == 0 {
		return mcp.NewToolResultText("No canvas files found"), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d canvas files:\n\n", len(canvases))
	for _, c := range canvases {
		fmt.Fprintf(&sb, "- %s\n", c)
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// ReadCanvasHandler reads and parses a canvas file
func (v *Vault) ReadCanvasHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	canvasPath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	if !strings.HasSuffix(canvasPath, ".canvas") {
		canvasPath += ".canvas"
	}

	fullPath := filepath.Join(v.path, canvasPath)
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Canvas not found: %s", canvasPath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read canvas: %v", err)), nil
	}

	var canvas Canvas
	if err := json.Unmarshal(content, &canvas); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid canvas format: %v", err)), nil
	}

	return mcp.NewToolResultText(formatCanvas(canvasPath, &canvas)), nil
}

// formatCanvas formats a canvas as markdown
func formatCanvas(path string, canvas *Canvas) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Canvas: %s\n\n", path)
	fmt.Fprintf(&sb, "**Nodes:** %d | **Edges:** %d\n\n", len(canvas.Nodes), len(canvas.Edges))

	// Group nodes by type
	byType := make(map[string][]*CanvasNode)
	for i := range canvas.Nodes {
		node := &canvas.Nodes[i]
		byType[node.Type] = append(byType[node.Type], node)
	}

	// Text nodes
	if texts := byType["text"]; len(texts) > 0 {
		fmt.Fprintf(&sb, "## Text Nodes (%d)\n", len(texts))
		for _, n := range texts {
			preview := n.Text
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			fmt.Fprintf(&sb, "- `%s`: %s\n", n.ID, preview)
		}
		sb.WriteString("\n")
	}

	// File nodes
	if files := byType["file"]; len(files) > 0 {
		fmt.Fprintf(&sb, "## File Nodes (%d)\n", len(files))
		for _, n := range files {
			fmt.Fprintf(&sb, "- `%s`: [[%s]]\n", n.ID, n.File)
		}
		sb.WriteString("\n")
	}

	// Link nodes
	if links := byType["link"]; len(links) > 0 {
		fmt.Fprintf(&sb, "## Link Nodes (%d)\n", len(links))
		for _, n := range links {
			fmt.Fprintf(&sb, "- `%s`: %s\n", n.ID, n.URL)
		}
		sb.WriteString("\n")
	}

	// Group nodes
	if groups := byType["group"]; len(groups) > 0 {
		fmt.Fprintf(&sb, "## Groups (%d)\n", len(groups))
		for _, n := range groups {
			label := n.Label
			if label == "" {
				label = "(unlabeled)"
			}
			fmt.Fprintf(&sb, "- `%s`: %s\n", n.ID, label)
		}
		sb.WriteString("\n")
	}

	// Edges
	if len(canvas.Edges) > 0 {
		fmt.Fprintf(&sb, "## Connections (%d)\n", len(canvas.Edges))
		for _, e := range canvas.Edges {
			label := ""
			if e.Label != "" {
				label = fmt.Sprintf(" [%s]", e.Label)
			}
			fmt.Fprintf(&sb, "- %s → %s%s\n", e.FromNode, e.ToNode, label)
		}
	}

	return sb.String()
}

// CreateCanvasHandler creates a new canvas file
func (v *Vault) CreateCanvasHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	canvasPath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	if !strings.HasSuffix(canvasPath, ".canvas") {
		canvasPath += ".canvas"
	}

	fullPath := filepath.Join(v.path, canvasPath)
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Check if exists
	if _, err := os.Stat(fullPath); err == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Canvas already exists: %s", canvasPath)), nil
	}

	// Create empty canvas
	canvas := Canvas{
		Nodes: []CanvasNode{},
		Edges: []CanvasEdge{},
	}

	data, err := json.MarshalIndent(canvas, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create canvas: %v", err)), nil
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	if err := os.WriteFile(fullPath, data, 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write canvas: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Created canvas: %s", canvasPath)), nil
}

// AddCanvasNodeHandler adds a node to a canvas
func (v *Vault) AddCanvasNodeHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	canvasPath, err := req.RequireString("canvas")
	if err != nil {
		return mcp.NewToolResultError("canvas path is required"), nil
	}

	nodeType := req.GetString("type", "text")
	content := req.GetString("content", "")
	x := int(req.GetInt("x", 0))
	y := int(req.GetInt("y", 0))
	width := int(req.GetInt("width", 300))
	height := int(req.GetInt("height", 200))

	if !strings.HasSuffix(canvasPath, ".canvas") {
		canvasPath += ".canvas"
	}

	fullPath := filepath.Join(v.path, canvasPath)
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Read existing canvas
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Canvas not found: %s", canvasPath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read canvas: %v", err)), nil
	}

	var canvas Canvas
	if err := json.Unmarshal(data, &canvas); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid canvas format: %v", err)), nil
	}

	// Generate unique ID
	nodeID := fmt.Sprintf("node-%d", len(canvas.Nodes)+1)

	// Create node based on type
	node := CanvasNode{
		ID:     nodeID,
		Type:   nodeType,
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}

	switch nodeType {
	case "text":
		node.Text = content
	case "file":
		node.File = content
	case "link":
		node.URL = content
	case "group":
		node.Label = content
	default:
		return mcp.NewToolResultError(fmt.Sprintf("Unknown node type: %s (use: text, file, link, group)", nodeType)), nil
	}

	canvas.Nodes = append(canvas.Nodes, node)

	// Write back
	newData, err := json.MarshalIndent(canvas, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize canvas: %v", err)), nil
	}

	if err := os.WriteFile(fullPath, newData, 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write canvas: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Added %s node `%s` to canvas", nodeType, nodeID)), nil
}

// AddCanvasEdgeHandler adds an edge between nodes in a canvas
func (v *Vault) AddCanvasEdgeHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	canvasPath, err := req.RequireString("canvas")
	if err != nil {
		return mcp.NewToolResultError("canvas path is required"), nil
	}

	fromNode, err := req.RequireString("from")
	if err != nil {
		return mcp.NewToolResultError("from node ID is required"), nil
	}

	toNode, err := req.RequireString("to")
	if err != nil {
		return mcp.NewToolResultError("to node ID is required"), nil
	}

	label := req.GetString("label", "")

	if !strings.HasSuffix(canvasPath, ".canvas") {
		canvasPath += ".canvas"
	}

	fullPath := filepath.Join(v.path, canvasPath)
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Read existing canvas
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Canvas not found: %s", canvasPath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read canvas: %v", err)), nil
	}

	var canvas Canvas
	if err := json.Unmarshal(data, &canvas); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid canvas format: %v", err)), nil
	}

	// Verify nodes exist
	nodeExists := func(id string) bool {
		for i := range canvas.Nodes {
			if canvas.Nodes[i].ID == id {
				return true
			}
		}
		return false
	}

	if !nodeExists(fromNode) {
		return mcp.NewToolResultError(fmt.Sprintf("From node not found: %s", fromNode)), nil
	}
	if !nodeExists(toNode) {
		return mcp.NewToolResultError(fmt.Sprintf("To node not found: %s", toNode)), nil
	}

	// Generate edge ID
	edgeID := fmt.Sprintf("edge-%d", len(canvas.Edges)+1)

	edge := CanvasEdge{
		ID:       edgeID,
		FromNode: fromNode,
		ToNode:   toNode,
		Label:    label,
	}

	canvas.Edges = append(canvas.Edges, edge)

	// Write back
	newData, err := json.MarshalIndent(canvas, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize canvas: %v", err)), nil
	}

	if err := os.WriteFile(fullPath, newData, 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write canvas: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Added edge `%s`: %s → %s", edgeID, fromNode, toNode)), nil
}
