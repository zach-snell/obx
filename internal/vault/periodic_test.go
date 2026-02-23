package vault

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestQuarterlyNoteHandlerUsesQuarterMonths(t *testing.T) {
	ctx := context.Background()
	v, _ := setupTestVault(t)

	result, _, err := v.QuarterlyNoteHandler(ctx, nil, QuarterlyNoteArgs{
		Date:            "2026-08-15",
		CreateIfMissing: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	for _, month := range []string{"[[2026-07]]", "[[2026-08]]", "[[2026-09]]"} {
		if !strings.Contains(text, month) {
			t.Fatalf("expected quarterly template to contain %s, got:\n%s", month, text)
		}
	}

	for _, wrong := range []string{"[[2026-01]]", "[[2026-02]]", "[[2026-03]]"} {
		if strings.Contains(text, wrong) {
			t.Fatalf("did not expect quarterly template to contain %s, got:\n%s", wrong, text)
		}
	}
}
