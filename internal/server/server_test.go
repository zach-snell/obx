package server

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerRegistration(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	// Wire up an in-memory client↔server session via the MCP protocol.
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Server must connect before client.
	if _, err := s.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	// List tools through the protocol — no reflection or unsafe needed.
	result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	toolMap := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		toolMap[tool.Name] = true
	}

	t.Logf("Found %d tools registered", len(result.Tools))

	if len(result.Tools) < 50 {
		t.Errorf("Expected at least 50 tools, got %d", len(result.Tools))
	}

	// Spot-check a representative tool from each category.
	expectedTools := []string{
		"list-notes",
		"read-note",
		"write-note",
		"edit-note",
		"batch-edit-note",
		"search-vault",
		"search-advanced",
		"daily-note",
		"list-tasks",
		"complete-tasks",
		"get-frontmatter",
		"query-frontmatter",
		"get-backlinks",
		"generate-moc",
		"bulk-tag",
		"list-templates",
		"create-canvas",
		"vault-stats",
	}

	for _, name := range expectedTools {
		if !toolMap[name] {
			t.Errorf("Expected tool %q not found in registered tools", name)
		}
	}
}
