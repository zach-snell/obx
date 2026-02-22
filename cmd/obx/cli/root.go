package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zach-snell/obx/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "obx",
	Short: "obx is the definitive CLI tool and MCP server for Obsidian",
	Long: `obx allows you to interact with your Obsidian vault from the command line 
or run it as a Model Context Protocol (MCP) server for AI assistants.

Set the OBSIDIAN_VAULT_PATH environment variable to avoid passing the vault path to every command.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Flags and generic setup goes here
}

// getVaultPath returns the vault path from args or environment variable
func getVaultPath(args []string) string {
	// If path is provided as the first argument to commands that accept it
	if len(args) > 0 {
		if _, err := os.Stat(args[0]); err == nil {
			return args[0]
		}
	}

	// Fall back to environment variable
	path := os.Getenv("OBSIDIAN_VAULT_PATH")

	// Fall back to config active vault
	if path == "" {
		cfg, err := config.Load()
		if err == nil && cfg.ActiveVault != "" {
			if activePath, ok := cfg.Vaults[cfg.ActiveVault]; ok {
				path = activePath
			}
		}
	}

	if path == "" {
		fmt.Fprintln(os.Stderr, "Error: Vault path not provided. Use 'obx vault add' to configure a vault, or set the OBSIDIAN_VAULT_PATH environment variable.")
		os.Exit(1)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Vault path does not exist: %s\n", path)
		os.Exit(1)
	}

	return path
}
