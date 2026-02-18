package vault

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestListNotesHandlerModeSwitch(t *testing.T) {
	ctx := context.Background()
	v, dir := setupTestVault(t)

	writeTestFile(t, dir, "one.md", "# one")
	writeTestFile(t, dir, "two.md", "# two")

	t.Run("compact default", func(t *testing.T) {
		result, _, err := v.ListNotesHandler(ctx, nil, ListNotesArgs{})
		if err != nil {
			t.Fatal(err)
		}

		text := result.Content[0].(*mcp.TextContent).Text
		var payload map[string]any
		if err := json.Unmarshal([]byte(text), &payload); err != nil {
			t.Fatalf("expected compact JSON response, got: %s", text)
		}
		if payload["mode"] != modeCompact {
			t.Fatalf("expected mode %q, got %v", modeCompact, payload["mode"])
		}

		data := payload["data"].(map[string]any)
		if int(data["total_count"].(float64)) != 2 {
			t.Fatalf("expected total_count 2, got %v", data["total_count"])
		}
	})

	t.Run("detailed mode", func(t *testing.T) {
		result, _, err := v.ListNotesHandler(ctx, nil, ListNotesArgs{Mode: modeDetailed})
		if err != nil {
			t.Fatal(err)
		}

		text := result.Content[0].(*mcp.TextContent).Text
		if strings.HasPrefix(text, "{") {
			t.Fatalf("expected detailed text response, got JSON: %s", text)
		}
		if !strings.Contains(text, "Found 2 notes") {
			t.Fatalf("unexpected detailed response: %s", text)
		}
	})
}

func TestSearchVaultHandlerModeSwitch(t *testing.T) {
	ctx := context.Background()
	v, dir := setupTestVault(t)

	writeTestFile(t, dir, "alpha.md", "hello world\nanother line")
	writeTestFile(t, dir, "beta.md", "hello from beta")

	t.Run("compact default", func(t *testing.T) {
		result, _, err := v.SearchVaultHandler(ctx, nil, SearchArgs{Query: "hello"})
		if err != nil {
			t.Fatal(err)
		}

		text := result.Content[0].(*mcp.TextContent).Text
		var payload map[string]any
		if err := json.Unmarshal([]byte(text), &payload); err != nil {
			t.Fatalf("expected compact JSON response, got: %s", text)
		}
		if payload["mode"] != modeCompact {
			t.Fatalf("expected mode %q, got %v", modeCompact, payload["mode"])
		}

		data := payload["data"].(map[string]any)
		if int(data["total_matches"].(float64)) != 2 {
			t.Fatalf("expected total_matches 2, got %v", data["total_matches"])
		}
	})

	t.Run("detailed mode", func(t *testing.T) {
		result, _, err := v.SearchVaultHandler(ctx, nil, SearchArgs{Query: "hello", Mode: modeDetailed})
		if err != nil {
			t.Fatal(err)
		}

		text := result.Content[0].(*mcp.TextContent).Text
		if strings.HasPrefix(text, "{") {
			t.Fatalf("expected detailed text response, got JSON: %s", text)
		}
		if !strings.Contains(text, `Found 2 matches for "hello"`) {
			t.Fatalf("unexpected detailed response: %s", text)
		}
	})
}
