package vault

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func writeBenchFile(tb testing.TB, root, relPath, content string) {
	tb.Helper()
	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		tb.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
		tb.Fatal(err)
	}
}

func extractTextResultSize(tb testing.TB, result *mcp.CallToolResult) int {
	tb.Helper()
	if len(result.Content) == 0 {
		return 0
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return 0
	}
	return len(text.Text)
}

func benchmarkListNotesByMode(b *testing.B, mode string) {
	ctx := context.Background()
	root := b.TempDir()
	v := New(root)

	for i := 0; i < 200; i++ {
		writeBenchFile(b, root, filepath.Join("bench", "note-"+strconv.Itoa(i)+".md"), "content")
	}

	args := ListNotesArgs{Mode: mode}
	b.ResetTimer()
	var bytes int
	for i := 0; i < b.N; i++ {
		result, _, err := v.ListNotesHandler(ctx, nil, args)
		if err != nil {
			b.Fatal(err)
		}
		bytes = extractTextResultSize(b, result)
	}
	b.ReportMetric(float64(bytes), "bytes/resp")
	b.ReportMetric(float64(bytes)/4.0, "tokens_est/resp")
}

func BenchmarkListNotesCompact(b *testing.B) {
	benchmarkListNotesByMode(b, modeCompact)
}

func BenchmarkListNotesDetailed(b *testing.B) {
	benchmarkListNotesByMode(b, modeDetailed)
}

func benchmarkSearchVaultByMode(b *testing.B, mode string) {
	ctx := context.Background()
	root := b.TempDir()
	v := New(root)

	for i := 0; i < 120; i++ {
		writeBenchFile(b, root, filepath.Join("search", "note-"+strconv.Itoa(i)+".md"), "alpha beta gamma token")
	}

	args := SearchArgs{Query: "alpha", Mode: mode}
	b.ResetTimer()
	var bytes int
	for i := 0; i < b.N; i++ {
		result, _, err := v.SearchVaultHandler(ctx, nil, args)
		if err != nil {
			b.Fatal(err)
		}
		bytes = extractTextResultSize(b, result)
	}
	b.ReportMetric(float64(bytes), "bytes/resp")
	b.ReportMetric(float64(bytes)/4.0, "tokens_est/resp")
}

func BenchmarkSearchVaultCompact(b *testing.B) {
	benchmarkSearchVaultByMode(b, modeCompact)
}

func BenchmarkSearchVaultDetailed(b *testing.B) {
	benchmarkSearchVaultByMode(b, modeDetailed)
}
