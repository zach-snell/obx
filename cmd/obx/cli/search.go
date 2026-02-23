package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/zach-snell/obx/internal/vault"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the vault for a given query",
	Long: `Search the Obsidian vault for text matching the given query.

By default, prints the results in a human-readable list. 
If the --json flag is provided, it returns structured, compact JSON 
perfect for piping to jq or other tools.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vaultPath := getVaultPath(nil)
		v := vault.New(vaultPath)
		ctx := context.Background()

		query := args[0]
		asJSON, _ := cmd.Flags().GetBool("json")

		mode := "detailed"
		if asJSON {
			mode = "compact"
		}

		res, _, err := v.SearchVaultHandler(ctx, nil, vault.SearchArgs{
			Query: query,
			Mode:  mode,
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Search failed: %v\n", err)
			os.Exit(1)
		}

		if len(res.Content) == 0 {
			if asJSON {
				fmt.Println("{}")
			} else {
				fmt.Println("No matches found.")
			}
			return
		}

		// The vault handler returns TextContent
		text := res.Content[0].(*mcp.TextContent).Text

		if asJSON {
			// In compact mode, the TextContent is actually a JSON string.
			// Let's verify it's valid JSON and format it if necessary, or just print it.
			var prettyJSON map[string]interface{}
			if err := json.Unmarshal([]byte(text), &prettyJSON); err == nil {
				formattedJSON, _ := json.MarshalIndent(prettyJSON, "", "  ")
				fmt.Println(string(formattedJSON))
			} else {
				// Fallback if not JSON for some reason
				fmt.Println(text)
			}
		} else {
			fmt.Println(text)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().Bool("json", false, "Output results in JSON format")
}
