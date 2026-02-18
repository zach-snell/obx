package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zach-snell/obsidian-go-mcp/internal/vault"
)

// New creates a new MCP server configured with vault tools
func New(vaultPath string) *mcp.Server {
	v := vault.New(vaultPath)

	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "Obsidian Vault MCP",
			Version: "0.3.3",
		},
		nil,
	)

	// Register tools
	registerTools(s, v)

	return s
}

func registerTools(s *mcp.Server, v *vault.Vault) {
	// Core Vault
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-notes",
		Description: "List all notes in the vault or a specific directory",
	}, v.ListNotesHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "read-note",
		Description: "Read the content of a specific note",
	}, v.ReadNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "write-note",
		Description: "Create or update a note",
	}, v.WriteNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-note",
		Description: "Delete a note",
	}, v.DeleteNoteHandler)

	// Editing
	mcp.AddTool(s, &mcp.Tool{
		Name:        "edit-note",
		Description: "Edit a section of a note using search and replace",
	}, v.EditNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "replace-section",
		Description: "Replace a section in a note identified by a heading",
	}, v.ReplaceSectionHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "append-note",
		Description: "Append content to a note",
	}, v.AppendNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "batch-edit-note",
		Description: "Apply multiple edits to a note in a single transaction",
	}, v.BatchEditNoteHandler)

	// Search
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-vault",
		Description: "Search for notes containing specific text",
	}, v.SearchVaultHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-advanced",
		Description: "Advanced search with boolean operators and options",
	}, v.SearchAdvancedHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-date",
		Description: "Search notes by creation or modification date",
	}, v.SearchDateHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-regex",
		Description: "Search notes using regular expressions",
	}, v.SearchRegexHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-tags",
		Description: "Search for notes with specific tags",
	}, v.SearchByTagsHandler)

	// Tasks
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-tasks",
		Description: "List tasks from notes",
	}, v.ListTasksHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "toggle-task",
		Description: "Toggle the status of a task",
	}, v.ToggleTaskHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "complete-tasks",
		Description: "Mark multiple tasks as complete by text match",
	}, v.CompleteTasksHandler)

	// Periodic Notes
	mcp.AddTool(s, &mcp.Tool{
		Name:        "daily-note",
		Description: "Get or create today's daily note",
	}, v.DailyNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-daily-notes",
		Description: "List recent daily notes",
	}, v.ListDailyNotesHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "weekly-note",
		Description: "Get or create this week's note",
	}, v.WeeklyNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "monthly-note",
		Description: "Get or create this month's note",
	}, v.MonthlyNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "quarterly-note",
		Description: "Get or create this quarter's note",
	}, v.QuarterlyNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "yearly-note",
		Description: "Get or create this year's note",
	}, v.YearlyNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-periodic-notes",
		Description: "List periodic notes",
	}, v.ListPeriodicNotesHandler)

	// Folders
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-folders",
		Description: "List all folders in the vault",
	}, v.ListFoldersHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-folder",
		Description: "Create a new folder",
	}, v.CreateFolderHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-folder",
		Description: "Delete a folder",
	}, v.DeleteFolderHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "move-note",
		Description: "Move a note to a new location",
	}, v.MoveNoteHandler)

	// Metadata / Frontmatter
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-frontmatter",
		Description: "Get frontmatter properties of a note",
	}, v.GetFrontmatterHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "query-frontmatter",
		Description: "Search notes by frontmatter properties",
	}, v.QueryFrontmatterHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "set-frontmatter",
		Description: "Set a frontmatter key",
	}, v.SetFrontmatterHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "remove-frontmatter",
		Description: "Remove a frontmatter key",
	}, v.RemoveFrontmatterKeyHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add-alias",
		Description: "Add an alias to a note",
	}, v.AddAliasHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add-tag",
		Description: "Add a tag to a note's frontmatter",
	}, v.AddTagToFrontmatterHandler)

	// Links & Backlinks
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-backlinks",
		Description: "Get backlinks for a note",
	}, v.BacklinksHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rename-note",
		Description: "Rename a note and update links",
	}, v.RenameNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "forward-links",
		Description: "Get forward links from a note",
	}, v.ForwardLinksHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "orphan-notes",
		Description: "Find orphan notes (no backlinks)",
	}, v.OrphanNotesHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "broken-links",
		Description: "Find broken links in the vault",
	}, v.BrokenLinksHandler)

	// Analysis
	mcp.AddTool(s, &mcp.Tool{
		Name:        "find-stubs",
		Description: "Find stub notes (short notes)",
	}, v.FindStubsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "find-outdated",
		Description: "Find outdated notes",
	}, v.FindOutdatedHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "unlinked-mentions",
		Description: "Find unlinked mentions of a note",
	}, v.UnlinkedMentionsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "suggest-links",
		Description: "Suggest links for a note based on content",
	}, v.SuggestLinksHandler)

	// Inline Fields (Dataview style)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-inline-fields",
		Description: "Get inline fields from a note",
	}, v.GetInlineFieldsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "set-inline-field",
		Description: "Set an inline field in a note",
	}, v.SetInlineFieldHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-inline-fields",
		Description: "Search/query notes by inline fields",
	}, v.QueryInlineFieldsHandler)

	// Bulk Operations
	mcp.AddTool(s, &mcp.Tool{
		Name:        "bulk-tag",
		Description: "Add or remove tags from multiple notes",
	}, v.BulkTagHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "bulk-move",
		Description: "Move multiple notes to a folder",
	}, v.BulkMoveHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "bulk-set-frontmatter",
		Description: "Set frontmatter on multiple notes",
	}, v.BulkSetFrontmatterHandler)

	// Refactoring
	mcp.AddTool(s, &mcp.Tool{
		Name:        "extract-note",
		Description: "Extract selected text to a new note",
	}, v.ExtractNoteHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "merge-notes",
		Description: "Merge multiple notes into one",
	}, v.MergeNotesHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "extract-section",
		Description: "Extract a section to a new note",
	}, v.ExtractSectionHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "duplicate-note",
		Description: "Duplicate a note",
	}, v.DuplicateNoteHandler)

	// Templates
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-templates",
		Description: "List available templates",
	}, v.ListTemplatesHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-template",
		Description: "Get template content",
	}, v.GetTemplateHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "apply-template",
		Description: "Apply a template to a note",
	}, v.ApplyTemplateHandler)

	// Canvas
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-canvases",
		Description: "List canvas files",
	}, v.ListCanvasesHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "read-canvas",
		Description: "Read a canvas file",
	}, v.ReadCanvasHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-canvas",
		Description: "Create a new canvas",
	}, v.CreateCanvasHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add-canvas-node",
		Description: "Add a node to a canvas",
	}, v.AddCanvasNodeHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add-canvas-edge",
		Description: "Add an edge (connection) to a canvas",
	}, v.AddCanvasEdgeHandler)

	// Maps of Content (MOC)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover-mocs",
		Description: "Discover potential MOCs",
	}, v.DiscoverMOCsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "generate-moc",
		Description: "Generate a Map of Content",
	}, v.GenerateMOCHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-moc",
		Description: "Update an existing MOC",
	}, v.UpdateMOCHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "generate-index",
		Description: "Generate an index note",
	}, v.GenerateIndexHandler)

	// Statistics
	mcp.AddTool(s, &mcp.Tool{
		Name:        "vault-stats",
		Description: "Get statistics about the vault",
	}, v.VaultStatsHandler)

	// Batch Read
	mcp.AddTool(s, &mcp.Tool{
		Name:        "read-notes",
		Description: "Read multiple notes at once",
	}, v.ReadNotesHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-note-summary",
		Description: "Get a summary of a note",
	}, v.GetNoteSummaryHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-section",
		Description: "Get a specific section of a note",
	}, v.GetSectionHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-headings",
		Description: "Get headings from a note",
	}, v.GetHeadingsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-headings",
		Description: "Search for headings across notes",
	}, v.SearchHeadingsHandler)
}
