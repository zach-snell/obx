package server

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerRegistration(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, []string{"search-vault"}, false, nil)

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

	if len(result.Tools) != 13 {
		t.Errorf("Expected 13 tools due to disable parameter, got %d", len(result.Tools))
	}

	if toolMap["search-vault"] {
		t.Errorf("Expected search-vault to be disabled")
	}

	expectedTools := []string{
		"manage-notes",
		"edit-note",
		"manage-periodic-notes",
		"manage-folders",
		"manage-frontmatter",
		"manage-tasks",
		"analyze-vault",
		"manage-canvas",
		"manage-mocs",
		"read-batch",
		"manage-links",
		"bulk-operations",
		"manage-templates",
	}

	for _, name := range expectedTools {
		if !toolMap[name] {
			t.Errorf("Expected tool %q not found in registered tools", name)
		}
	}
}

func TestManageVaultsRegistration(t *testing.T) {
	dir := t.TempDir()

	// Enable vault switching
	s := New(dir, nil, true, map[string]string{"test": dir})

	// Wire up an in-memory client↔server session via the MCP protocol.
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := s.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "Test Client",
		Version: "1.0.0",
	}, nil)

	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	// Request tool list
	result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	toolMap := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		toolMap[tool.Name] = true
	}

	if len(result.Tools) != 15 {
		t.Errorf("Expected 15 tools with vault switching enabled, got %d", len(result.Tools))
	}

	if !toolMap["manage-vaults"] {
		t.Errorf("Expected 'manage-vaults' to be registered when enabled")
	}
}
