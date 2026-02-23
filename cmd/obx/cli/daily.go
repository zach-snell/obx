package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/zach-snell/obx/internal/vault"
)

var dailyCmd = &cobra.Command{
	Use:   "daily [optional entry text]",
	Short: "Get today's daily note, optionally appending text to it",
	Long: `Retrieve or create today's daily note.

If text is provided as arguments, it will be automatically appended to the Daily Note.
This makes for a powerful, lightning-fast quick-capture tool from your terminal.

Examples:
  obx daily
  obx daily "Just had a great idea for the new project"`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultPath := getVaultPath(nil) // The text args are not a vault path
		v := vault.New(vaultPath)
		ctx := context.Background()

		// 1. Get or create the daily note
		res, _, err := v.DailyNoteHandler(ctx, nil, vault.DailyNoteArgs{
			CreateIfMissing: true,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(res.Content) == 0 {
			fmt.Println("Error: No content returned for daily note")
			return
		}

		text := res.Content[0].(*mcp.TextContent).Text

		// If we only wanted to open/create it, we're done
		if len(args) == 0 {
			fmt.Println(text)
			return
		}

		// 2. We have text to append! Let's append to it.
		// First we must extract the file path from the "Created: path/to.md" or "Path: path/to.md" output.
		lines := strings.Split(text, "\n")
		pathLine := lines[0]

		var notePath string
		if strings.HasPrefix(pathLine, "Created: ") {
			notePath = strings.TrimPrefix(pathLine, "Created: ")
		} else if strings.HasPrefix(pathLine, "Path: ") {
			notePath = strings.TrimPrefix(pathLine, "Path: ")
		} else {
			fmt.Printf("Error parsing note path from output: %s\n", pathLine)
			return
		}
		notePath = strings.TrimSpace(notePath)

		// Clean up the text we want to append
		entryText := strings.Join(args, " ")

		// 3. Append to the note
		appendRes, _, err := v.AppendNoteHandler(ctx, nil, vault.AppendNoteArgs{
			Path:    notePath,
			Content: "- " + entryText, // Format as a bullet point by default
		})

		if err != nil {
			fmt.Printf("Error appending to %s: %v\n", notePath, err)
			return
		}

		if len(appendRes.Content) > 0 {
			fmt.Println(appendRes.Content[0].(*mcp.TextContent).Text)
		}
	},
}

func init() {
	rootCmd.AddCommand(dailyCmd)
}
