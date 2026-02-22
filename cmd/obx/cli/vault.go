package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zach-snell/obx/internal/config"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage configured vaults",
	Long:  `Add, list, or switch between configured Obsidian vaults.`,
}

var vaultAddCmd = &cobra.Command{
	Use:   "add [alias] [path]",
	Short: "Add a new vault configuration",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		path := args[1]

		// Get absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			return
		}

		// Verify directory exists
		if info, err := os.Stat(absPath); err != nil || !info.IsDir() {
			fmt.Printf("Error: Path does not exist or is not a directory: %s\n", absPath)
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		cfg.Vaults[alias] = absPath

		// If this is the first vault, make it active
		if cfg.ActiveVault == "" {
			cfg.ActiveVault = alias
			fmt.Printf("Set '%s' as the active vault.\n", alias)
		}

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Successfully added vault '%s' at '%s'\n", alias, absPath)
	},
}

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured vaults",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if len(cfg.Vaults) == 0 {
			fmt.Println("No vaults configured. Use 'obx vault add <alias> <path>' to add one.")
			return
		}

		fmt.Println("Configured Vaults:")
		for alias, path := range cfg.Vaults {
			activeMark := " "
			if alias == cfg.ActiveVault {
				activeMark = "*"
			}
			fmt.Printf("%s %s: %s\n", activeMark, alias, path)
		}
	},
}

var vaultSwitchCmd = &cobra.Command{
	Use:   "switch [alias]",
	Short: "Switch the active vault",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if _, exists := cfg.Vaults[alias]; !exists {
			fmt.Printf("Error: Vault '%s' not found.\n", alias)
			return
		}

		cfg.ActiveVault = alias
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Switched active vault to '%s'\n", alias)
	},
}

var vaultRemoveCmd = &cobra.Command{
	Use:   "remove [alias]",
	Short: "Remove a vault configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if _, exists := cfg.Vaults[alias]; !exists {
			fmt.Printf("Error: Vault '%s' not found.\n", alias)
			return
		}

		delete(cfg.Vaults, alias)

		// Unset active if we deleted the active one
		if cfg.ActiveVault == alias {
			cfg.ActiveVault = ""

			// Auto-pick another if one exists
			for k := range cfg.Vaults {
				cfg.ActiveVault = k
				fmt.Printf("Removed active vault. Defaulted active vault to '%s'\n", k)
				break
			}
		}

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Removed vault '%s'\n", alias)
	},
}

func init() {
	vaultCmd.AddCommand(vaultAddCmd)
	vaultCmd.AddCommand(vaultListCmd)
	vaultCmd.AddCommand(vaultSwitchCmd)
	vaultCmd.AddCommand(vaultRemoveCmd)
	rootCmd.AddCommand(vaultCmd)
}
