package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/zach-snell/obx/internal/vault"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check your vault for broken links and orphan notes",
	Long: `Analyze the vault for common issues like broken wikilinks or notes 
that are not linked from any other note.`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultPath := getVaultPath(args)
		v := vault.New(vaultPath)
		ctx := context.Background()

		fmt.Println("Running vault health checks...")
		fmt.Printf("Vault path: %s\n\n", vaultPath)

		hasIssues := false

		// 1. Check for Broken Links
		fmt.Println("--- Broken Links ---")
		blRes, _, err := v.BrokenLinksHandler(ctx, nil, vault.BrokenLinksArgs{})
		if err != nil {
			fmt.Printf("Error checking broken links: %v\n", err)
			hasIssues = true
		} else if len(blRes.Content) > 0 {
			text := blRes.Content[0].(*mcp.TextContent).Text
			fmt.Println(text)
			if text != "No broken links found" {
				hasIssues = true
			}
		}

		fmt.Println("\n--- Orphan Notes ---")
		// 2. Check for Orphan Notes
		onRes, _, err := v.OrphanNotesHandler(ctx, nil, vault.OrphanNotesArgs{})
		if err != nil {
			fmt.Printf("Error checking orphan notes: %v\n", err)
			hasIssues = true
		} else if len(onRes.Content) > 0 {
			text := onRes.Content[0].(*mcp.TextContent).Text
			fmt.Println(text)
			if text != "No orphan notes found" {
				hasIssues = true
			}
		}

		failOnIssues, _ := cmd.Flags().GetBool("fail")
		if failOnIssues && hasIssues {
			fmt.Println("\nDoctor found issues. Exiting with failure code.")
			os.Exit(1)
		} else if hasIssues {
			fmt.Println("\nDoctor found issues. Use --fail to exit with an error code.")
		} else {
			fmt.Println("\nVault looks healthy! ✨")
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().Bool("fail", false, "Exit with non-zero status if issues are found")
}
