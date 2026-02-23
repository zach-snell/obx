package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zach-snell/obx/internal/vault"
)

// New creates a new MCP server configured with vault tools
func New(vaultPath string, disabledTools []string, allowVaultSwitching bool, allowedVaults map[string]string) *mcp.Server {
	v := vault.New(vaultPath)
	if allowedVaults != nil {
		v.SetAllowedVaults(allowedVaults)
	}

	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "Obsidian Vault MCP",
			Version: "0.3.3",
		},
		nil,
	)

	// Register tools
	registerTools(s, v, disabledTools, allowVaultSwitching)

	return s
}

func isToolDisabled(name string, disabledTools []string) bool {
	for _, t := range disabledTools {
		if t == name {
			return true
		}
	}
	return false
}

func registerTools(s *mcp.Server, v *vault.Vault, disabledTools []string, allowVaultSwitching bool) {
	if allowVaultSwitching && !isToolDisabled("manage-vaults", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-vaults",
			Description: "List and switch between available Obsidian vaults dynamically",
		}, v.ManageVaultsMultiplexHandler)
	}

	if !isToolDisabled("manage-notes", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-notes",
			Description: "Unified tool for listing, reading, writing, moving, deleting, renaming, and appending to notes",
		}, v.ManageNotesMultiplexHandler)
	}

	if !isToolDisabled("edit-note", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "edit-note",
			Description: "Unified tool for targeted text edits, section replacements, and batch edits",
		}, v.EditNoteMultiplexHandler)
	}

	if !isToolDisabled("search-vault", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "search-vault",
			Description: "Unified search tool spanning text query, advanced block search, regex, dates, tags, inline-fields, and frontmatter queries",
		}, v.SearchVaultMultiplexHandler)
	}

	if !isToolDisabled("manage-periodic-notes", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-periodic-notes",
			Description: "Unified tool for getting, creating, and listing daily, weekly, monthly, and yearly periodic notes",
		}, v.ManagePeriodicNotesMultiplexHandler)
	}

	if !isToolDisabled("manage-folders", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-folders",
			Description: "Unified tool for listing, creating, and deleting folders",
		}, v.ManageFoldersMultiplexHandler)
	}

	if !isToolDisabled("manage-frontmatter", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-frontmatter",
			Description: "Unified tool for manipulating note frontmatter properties, tags, aliases, and inline fields",
		}, v.ManageFrontmatterMultiplexHandler)
	}

	if !isToolDisabled("manage-tasks", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-tasks",
			Description: "Unified tool for finding, toggling, and completing checkbox tasks across the vault",
		}, v.ManageTasksMultiplexHandler)
	}

	if !isToolDisabled("analyze-vault", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "analyze-vault",
			Description: "Unified analytical tool to get vault stats, detect broken links, orphans, stubs, and outdated notes",
		}, v.AnalyzeVaultMultiplexHandler)
	}

	if !isToolDisabled("manage-canvas", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-canvas",
			Description: "Unified tool for reading, creating, and interacting with Canvas notes",
		}, v.ManageCanvasMultiplexHandler)
	}

	if !isToolDisabled("manage-mocs", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-mocs",
			Description: "Unified tool to discover and generate Maps of Content (MOCs) and folder indices",
		}, v.ManageMocsMultiplexHandler)
	}

	if !isToolDisabled("read-batch", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "read-batch",
			Description: "Unified tool for reading batches of notes, specific headings, sections, or generated summaries",
		}, v.ReadBatchMultiplexHandler)
	}

	if !isToolDisabled("manage-links", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-links",
			Description: "Unified tool covering backlinks, forward-links, and AI link suggestions for notes",
		}, v.ManageLinksMultiplexHandler)
	}

	if !isToolDisabled("bulk-operations", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "bulk-operations",
			Description: "Unified bulk operational tool for tagging, moving, and updating frontmatter across multiple notes",
		}, v.BulkOperationsMultiplexHandler)
	}

	if !isToolDisabled("manage-templates", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "manage-templates",
			Description: "Unified tool for listing, retrieving, and applying markdown templates",
		}, v.ManageTemplatesMultiplexHandler)
	}

	if !isToolDisabled("refactor-notes", disabledTools) {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "refactor-notes",
			Description: "Unified tool for structural note refactoring: split notes by heading, merge multiple notes, and extract sections to new notes",
		}, v.RefactorNotesMultiplexHandler)
	}
}
