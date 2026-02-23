package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/zach-snell/obx/internal/config"
	mcpserver "github.com/zach-snell/obx/internal/server"
)

var serveCmd = &cobra.Command{
	Use:     "mcp [vault-path]",
	Aliases: []string{"serve"},
	Short:   "Start the MCP server",
	Long: `Start the Obsidian MCP server to allow AI assistants to 
interact with your vault. Uses stdio transport by default, or HTTP streamable
if --http is provided.`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultPath := getVaultPath(args)
		disabledTools, _ := cmd.Flags().GetStringSlice("disabled-tools")

		allowSwitching, _ := cmd.Flags().GetBool("allow-vault-switching")
		allowedVaultsFlag, _ := cmd.Flags().GetStringSlice("allowed-vaults")

		var allowedVaults map[string]string
		if allowSwitching {
			allowedVaults = make(map[string]string)
			cfg, err := config.Load()
			if err == nil && cfg.Vaults != nil {
				if len(allowedVaultsFlag) > 0 {
					for _, alias := range allowedVaultsFlag {
						if path, ok := cfg.Vaults[alias]; ok {
							allowedVaults[alias] = path
						}
					}
				} else {
					allowedVaults = cfg.Vaults
				}
			}
		}

		// Create and configure MCP server
		s := mcpserver.New(vaultPath, disabledTools, allowSwitching, allowedVaults)

		// Determine transport
		addr, _ := cmd.Flags().GetString("http")
		if addr == "" && os.Getenv("OBSIDIAN_ADDR") != "" {
			addr = os.Getenv("OBSIDIAN_ADDR")
		}

		if addr != "" {
			serveHTTP(s, vaultPath, addr)
		} else {
			serveStdio(s, vaultPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().String("http", "", "HTTP listen address (e.g., :8080)")
	serveCmd.Flags().StringSlice("disabled-tools", []string{}, "Comma-separated list of unified tools to disable (e.g., manage-folders,bulk-operations)")
	serveCmd.Flags().Bool("allow-vault-switching", false, "Expose the manage-vaults MCP tool to allow agents to switch the active vault")
	serveCmd.Flags().StringSlice("allowed-vaults", []string{}, "Optional comma-separated list of vault aliases an agent is allowed to switch to. If empty but switching is enabled, all vaults are allowed.")
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
