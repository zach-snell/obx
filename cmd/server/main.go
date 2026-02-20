package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpserver "github.com/zach-snell/obsidian-go-mcp/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: obsidian-mcp <vault-path> [--http <addr>]")
	}

	vaultPath := os.Args[1]

	// Verify vault path exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		log.Fatalf("Vault path does not exist: %s", vaultPath)
	}

	// Create and configure MCP server
	s := mcpserver.New(vaultPath)

	// Determine transport: HTTP streamable or stdio
	addr := httpAddr()

	if addr != "" {
		serveHTTP(s, vaultPath, addr)
	} else {
		serveStdio(s, vaultPath)
	}
}

// httpAddr returns the HTTP listen address from --http flag or OBSIDIAN_ADDR env var.
// Returns "" if stdio transport should be used.
func httpAddr() string {
	// Check --http flag
	for i, arg := range os.Args {
		if arg == "--http" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}

	// Check env var
	return os.Getenv("OBSIDIAN_ADDR")
}

func serveStdio(s *mcp.Server, vaultPath string) {
	fmt.Fprintf(os.Stderr, "Starting Obsidian MCP Server (stdio)...\n")
	fmt.Fprintf(os.Stderr, "Vault: %s\n", vaultPath)

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func serveHTTP(s *mcp.Server, vaultPath, addr string) {
	fmt.Fprintf(os.Stderr, "Starting Obsidian MCP Server (HTTP Streamable)...\n")
	fmt.Fprintf(os.Stderr, "Vault: %s\n", vaultPath)
	fmt.Fprintf(os.Stderr, "Listening on %s\n", addr)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s
	}, nil)

	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
