package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerRegistration(t *testing.T) {
	// Create a temporary directory for the vault
	dir := t.TempDir()

	// Initialize the server
	var s *mcp.Server
	s = New(dir)

	// Use reflection to access private 'tools' field in mcp.Server
	val := reflect.ValueOf(s).Elem()
	toolsField := val.FieldByName("tools")

	if !toolsField.IsValid() {
		t.Fatal("Could not find 'tools' field in mcp.Server")
	}

	// Make sure we can access unexported field
	realTools := reflect.NewAt(toolsField.Type(), unsafe.Pointer(toolsField.UnsafeAddr())).Elem()

	// Dereference if pointer
	if realTools.Kind() == reflect.Ptr {
		realTools = realTools.Elem()
	}

	if realTools.Kind() != reflect.Struct {
		t.Fatalf("Expected 'tools' to be a struct, got %v", realTools.Kind())
	}

	// Find the 'features' map field
	// Based on debug output, it is 'features' (lowercase)
	featuresField := realTools.FieldByName("features")
	if !featuresField.IsValid() {
		// Try exported 'Features' just in case? No, likely 'features'
		t.Fatal("Could not find 'features' field in tools struct")
	}

	// Make addressable if unexported
	featuresMap := reflect.NewAt(featuresField.Type(), unsafe.Pointer(featuresField.UnsafeAddr())).Elem()

	if featuresMap.Kind() != reflect.Map {
		t.Fatalf("Expected 'features' to be a map, got %v", featuresMap.Kind())
	}

	count := featuresMap.Len()
	t.Logf("Found %d tools registered", count)

	// Verify tool count (expecting around 57)
	if count < 50 {
		t.Errorf("Expected at least 50 tools, got %d", count)
	}

	// Verify expected tools
	expectedTools := []string{
		"list-notes",
		"read-note",
		"write-note",
		"daily-note",
		"search-vault",
		"generate-moc",
		"get-frontmatter",
	}

	keys := featuresMap.MapKeys()
	toolMap := make(map[string]bool)
	for _, k := range keys {
		// The key is string
		if k.Kind() == reflect.String {
			toolMap[k.String()] = true
		}
	}

	for _, name := range expectedTools {
		if !toolMap[name] {
			t.Errorf("Expected tool %q not found", name)
		}
	}
}
