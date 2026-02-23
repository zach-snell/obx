package vault

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ManageNotesMultiplexArgs multiplexed args
type ManageNotesMultiplexArgs struct {
	Action        string `json:"action" jsonschema:"Action to perform: 'read', 'write', 'delete', 'append', 'rename', 'duplicate', 'move', 'list'"`
	Path          string `json:"path,omitempty" jsonschema:"Path to the note relative to vault root"`
	Content       string `json:"content,omitempty" jsonschema:"Content of the note"`
	ExpectedMtime string `json:"expected_mtime,omitempty" jsonschema:"Expected file modification time (RFC3339Nano) for optimistic concurrency"`
	DryRun        bool   `json:"dry_run,omitempty" jsonschema:"Preview deletion without modifying files"`
	Position      string `json:"position,omitempty" jsonschema:"Position to insert: 'end' (default), 'start', 'before', 'after'"`
	After         string `json:"after,omitempty" jsonschema:"Heading or text to insert after (if position is 'after')"`
	Before        string `json:"before,omitempty" jsonschema:"Heading or text to insert before (if position is 'before')"`
	ContextLines  int    `json:"context_lines,omitempty" jsonschema:"Number of context lines to return (default 0)"`
	OldPath       string `json:"old_path,omitempty" jsonschema:"Old note path"`
	NewPath       string `json:"new_path,omitempty" jsonschema:"New note path"`
	Output        string `json:"output,omitempty" jsonschema:"Output note path"`
	Source        string `json:"source,omitempty" jsonschema:"Source path"`
	Destination   string `json:"destination,omitempty" jsonschema:"Destination path"`
	UpdateLinks   bool   `json:"update_links,omitempty" jsonschema:"Whether to update links to this file (default true)"`
	Directory     string `json:"directory,omitempty" jsonschema:"Directory path relative to vault root (for list action)"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum number of notes to return (for list action, 0 = no limit)"`
	Offset        int    `json:"offset,omitempty" jsonschema:"Number of notes to skip for pagination (for list action, default 0)"`
	Mode          string `json:"mode,omitempty" jsonschema:"Response mode: compact (default) or detailed"`
}

// ManageNotesMultiplexHandler routes to the specific handler
func (v *Vault) ManageNotesMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageNotesMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "read":
		specificArgs := ReadNoteArgs{
			Path: args.Path,
		}
		return v.ReadNoteHandler(ctx, req, specificArgs)
	case "write":
		specificArgs := WriteNoteArgs{
			Path:          args.Path,
			Content:       args.Content,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.WriteNoteHandler(ctx, req, specificArgs)
	case "delete":
		specificArgs := DeleteNoteArgs{
			Path:          args.Path,
			DryRun:        args.DryRun,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.DeleteNoteHandler(ctx, req, specificArgs)
	case "append":
		specificArgs := AppendNoteArgs{
			Path:          args.Path,
			Content:       args.Content,
			Position:      args.Position,
			After:         args.After,
			Before:        args.Before,
			ContextLines:  args.ContextLines,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.AppendNoteHandler(ctx, req, specificArgs)
	case "rename":
		specificArgs := RenameNoteArgs{
			OldPath: args.OldPath,
			NewPath: args.NewPath,
		}
		return v.RenameNoteHandler(ctx, req, specificArgs)
	case "duplicate":
		specificArgs := DuplicateNoteArgs{
			Path:   args.Path,
			Output: args.Output,
		}
		return v.DuplicateNoteHandler(ctx, req, specificArgs)
	case "move":
		specificArgs := MoveArgs{
			Source:      args.Source,
			Destination: args.Destination,
			UpdateLinks: args.UpdateLinks,
			DryRun:      args.DryRun,
		}
		return v.MoveNoteHandler(ctx, req, specificArgs)
	case "list":
		specificArgs := ListNotesArgs{
			Directory: args.Directory,
			Limit:     args.Limit,
			Offset:    args.Offset,
			Mode:      args.Mode,
		}
		return v.ListNotesHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// EditNoteMultiplexArgs multiplexed args
type EditNoteMultiplexArgs struct {
	Action        string      `json:"action" jsonschema:"Action to perform: 'edit', 'replace-section', 'batch-edit'"`
	Path          string      `json:"path,omitempty" jsonschema:"Path to the note"`
	OldText       string      `json:"old_text,omitempty" jsonschema:"Text to find and replace"`
	NewText       string      `json:"new_text,omitempty" jsonschema:"Replacement text"`
	ReplaceAll    bool        `json:"replace_all,omitempty" jsonschema:"Whether to replace all occurrences (default false)"`
	ContextLines  int         `json:"context_lines,omitempty" jsonschema:"Number of context lines to return (default 0)"`
	ExpectedMtime string      `json:"expected_mtime,omitempty" jsonschema:"Expected file modification time (RFC3339Nano) for optimistic concurrency"`
	Heading       string      `json:"heading,omitempty" jsonschema:"Heading of the section to replace"`
	Content       string      `json:"content,omitempty" jsonschema:"New content for the section"`
	Edits         []EditEntry `json:"edits,omitempty" jsonschema:"List of edits to apply"`
	DryRun        bool        `json:"dry_run,omitempty" jsonschema:"Preview edits without modifying files"`
}

// EditNoteMultiplexHandler routes to the specific handler
func (v *Vault) EditNoteMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args EditNoteMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "edit":
		specificArgs := EditNoteArgs{
			Path:          args.Path,
			OldText:       args.OldText,
			NewText:       args.NewText,
			ReplaceAll:    args.ReplaceAll,
			ContextLines:  args.ContextLines,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.EditNoteHandler(ctx, req, specificArgs)
	case "replace-section":
		specificArgs := ReplaceSectionArgs{
			Path:          args.Path,
			Heading:       args.Heading,
			Content:       args.Content,
			ContextLines:  args.ContextLines,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.ReplaceSectionHandler(ctx, req, specificArgs)
	case "batch-edit":
		specificArgs := BatchEditArgs{
			Path:          args.Path,
			Edits:         args.Edits,
			ContextLines:  args.ContextLines,
			DryRun:        args.DryRun,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.BatchEditNoteHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// SearchVaultMultiplexArgs multiplexed args
type SearchVaultMultiplexArgs struct {
	Action          string `json:"action" jsonschema:"Action to perform: 'search', 'advanced', 'date', 'regex', 'tags', 'headings', 'inline-fields', 'frontmatter'"`
	Query           string `json:"query,omitempty" jsonschema:"Search query"`
	Directory       string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	Mode            string `json:"mode,omitempty" jsonschema:"Response mode: compact (default) or detailed"`
	SearchIn        string `json:"in,omitempty" jsonschema:"Where to search: 'content' (default), 'file', 'heading', 'block'"`
	Operator        string `json:"operator,omitempty" jsonschema:"Logical operator: 'and' (default), 'or'"`
	Limit           int    `json:"limit,omitempty" jsonschema:"Maximum results to return (default 50)"`
	From            string `json:"from,omitempty" jsonschema:"Start date (YYYY-MM-DD)"`
	To              string `json:"to,omitempty" jsonschema:"End date (YYYY-MM-DD)"`
	DateType        string `json:"type,omitempty" jsonschema:"Date type to check: 'modified' (default), 'created'"`
	Pattern         string `json:"pattern,omitempty" jsonschema:"Regex pattern"`
	CaseInsensitive bool   `json:"case_insensitive,omitempty" jsonschema:"Whether to ignore case (default true)"`
	Tags            string `json:"tags,omitempty" jsonschema:"Comma-separated list of tags"`
	Level           int    `json:"level,omitempty" jsonschema:"Heading level to filter (0 for all)"`
	Key             string `json:"key,omitempty" jsonschema:"Field key"`
	Value           string `json:"value,omitempty" jsonschema:"Field value to match (optional)"`
}

// SearchVaultMultiplexHandler routes to the specific handler
func (v *Vault) SearchVaultMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args SearchVaultMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "search":
		specificArgs := SearchArgs{
			Query:     args.Query,
			Directory: args.Directory,
			Mode:      args.Mode,
		}
		return v.SearchVaultHandler(ctx, req, specificArgs)
	case "advanced":
		specificArgs := SearchAdvancedArgs{
			Query:     args.Query,
			SearchIn:  args.SearchIn,
			Operator:  args.Operator,
			Directory: args.Directory,
			Limit:     args.Limit,
			Mode:      args.Mode,
		}
		return v.SearchAdvancedHandler(ctx, req, specificArgs)
	case "date":
		specificArgs := SearchDateArgs{
			From:      args.From,
			To:        args.To,
			DateType:  args.DateType,
			Directory: args.Directory,
			Limit:     args.Limit,
		}
		return v.SearchDateHandler(ctx, req, specificArgs)
	case "regex":
		specificArgs := SearchRegexArgs{
			Pattern:         args.Pattern,
			Directory:       args.Directory,
			Limit:           args.Limit,
			CaseInsensitive: args.CaseInsensitive,
			Mode:            args.Mode,
		}
		return v.SearchRegexHandler(ctx, req, specificArgs)
	case "tags":
		specificArgs := SearchTagsArgs{
			Tags:      args.Tags,
			Directory: args.Directory,
		}
		return v.SearchByTagsHandler(ctx, req, specificArgs)
	case "headings":
		specificArgs := SearchHeadingsArgs{
			Query:     args.Query,
			Level:     args.Level,
			Directory: args.Directory,
		}
		return v.SearchHeadingsHandler(ctx, req, specificArgs)
	case "inline-fields":
		specificArgs := SearchInlineFieldsArgs{
			Key:       args.Key,
			Value:     args.Value,
			Operator:  args.Operator,
			Directory: args.Directory,
		}
		return v.QueryInlineFieldsHandler(ctx, req, specificArgs)
	case "frontmatter":
		specificArgs := QueryFrontmatterArgs{
			Query:     args.Query,
			Directory: args.Directory,
		}
		return v.QueryFrontmatterHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManagePeriodicNotesMultiplexArgs multiplexed args
type ManagePeriodicNotesMultiplexArgs struct {
	Action          string `json:"action" jsonschema:"Action to perform: 'daily', 'weekly', 'monthly', 'quarterly', 'yearly', 'list-daily', 'list-periodic'"`
	Date            string `json:"date,omitempty" jsonschema:"Date string (default: today)"`
	Folder          string `json:"folder,omitempty" jsonschema:"Folder for daily notes (default: 'daily')"`
	Format          string `json:"format,omitempty" jsonschema:"Date format (default: '2006-01-02')"`
	CreateIfMissing bool   `json:"create,omitempty" jsonschema:"Create if missing (default: true)"`
	Type            string `json:"type,omitempty" jsonschema:"Type of note: 'daily', 'weekly', 'monthly', 'quarterly', 'yearly'"`
	Limit           int    `json:"limit,omitempty" jsonschema:"Maximum number of notes to return"`
}

// ManagePeriodicNotesMultiplexHandler routes to the specific handler
func (v *Vault) ManagePeriodicNotesMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManagePeriodicNotesMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "daily":
		specificArgs := DailyNoteArgs{
			Date:            args.Date,
			Folder:          args.Folder,
			Format:          args.Format,
			CreateIfMissing: args.CreateIfMissing,
		}
		return v.DailyNoteHandler(ctx, req, specificArgs)
	case "weekly":
		specificArgs := WeeklyNoteArgs{
			Date:            args.Date,
			Folder:          args.Folder,
			Format:          args.Format,
			CreateIfMissing: args.CreateIfMissing,
		}
		return v.WeeklyNoteHandler(ctx, req, specificArgs)
	case "monthly":
		specificArgs := MonthlyNoteArgs{
			Date:            args.Date,
			Folder:          args.Folder,
			Format:          args.Format,
			CreateIfMissing: args.CreateIfMissing,
		}
		return v.MonthlyNoteHandler(ctx, req, specificArgs)
	case "quarterly":
		specificArgs := QuarterlyNoteArgs{
			Date:            args.Date,
			Folder:          args.Folder,
			CreateIfMissing: args.CreateIfMissing,
		}
		return v.QuarterlyNoteHandler(ctx, req, specificArgs)
	case "yearly":
		specificArgs := YearlyNoteArgs{
			Date:            args.Date,
			Folder:          args.Folder,
			CreateIfMissing: args.CreateIfMissing,
		}
		return v.YearlyNoteHandler(ctx, req, specificArgs)
	case "list-daily":
		specificArgs := ListPeriodicArgs{
			Type:   args.Type,
			Limit:  args.Limit,
			Folder: args.Folder,
		}
		return v.ListDailyNotesHandler(ctx, req, specificArgs)
	case "list-periodic":
		specificArgs := ListPeriodicArgs{
			Type:   args.Type,
			Limit:  args.Limit,
			Folder: args.Folder,
		}
		return v.ListPeriodicNotesHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManageFoldersMultiplexArgs multiplexed args
type ManageFoldersMultiplexArgs struct {
	Action       string `json:"action" jsonschema:"Action to perform: 'list', 'create', 'delete'"`
	Directory    string `json:"directory,omitempty" jsonschema:"Root directory to list from"`
	IncludeEmpty bool   `json:"include_empty,omitempty" jsonschema:"Whether to include empty directories (default true)"`
	Path         string `json:"path,omitempty" jsonschema:"Path of the directory to create"`
	Force        bool   `json:"force,omitempty" jsonschema:"Force delete even if not empty (default false)"`
	DryRun       bool   `json:"dry_run,omitempty" jsonschema:"Preview deletion without modifying files"`
}

// ManageFoldersMultiplexHandler routes to the specific handler
func (v *Vault) ManageFoldersMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageFoldersMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "list":
		specificArgs := ListDirsArgs{
			Directory:    args.Directory,
			IncludeEmpty: args.IncludeEmpty,
		}
		return v.ListFoldersHandler(ctx, req, specificArgs)
	case "create":
		specificArgs := CreateDirArgs{
			Path: args.Path,
		}
		return v.CreateFolderHandler(ctx, req, specificArgs)
	case "delete":
		specificArgs := DeleteDirArgs{
			Path:   args.Path,
			Force:  args.Force,
			DryRun: args.DryRun,
		}
		return v.DeleteFolderHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManageFrontmatterMultiplexArgs multiplexed args
type ManageFrontmatterMultiplexArgs struct {
	Action        string `json:"action" jsonschema:"Action to perform: 'get', 'set', 'remove', 'add-alias', 'add-tag', 'get-inline-fields', 'set-inline-field'"`
	Path          string `json:"path,omitempty" jsonschema:"Path to the note"`
	Key           string `json:"key,omitempty" jsonschema:"Frontmatter key"`
	Value         string `json:"value,omitempty" jsonschema:"Value to set"`
	ExpectedMtime string `json:"expected_mtime,omitempty" jsonschema:"Expected file modification time (RFC3339Nano) for optimistic concurrency"`
	Alias         string `json:"alias,omitempty" jsonschema:"Alias to add"`
	Tag           string `json:"tag,omitempty" jsonschema:"Tag to add"`
}

// ManageFrontmatterMultiplexHandler routes to the specific handler
func (v *Vault) ManageFrontmatterMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageFrontmatterMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "get":
		specificArgs := GetFrontmatterArgs{
			Path: args.Path,
		}
		return v.GetFrontmatterHandler(ctx, req, specificArgs)
	case "set":
		specificArgs := SetFrontmatterArgs{
			Path:          args.Path,
			Key:           args.Key,
			Value:         args.Value,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.SetFrontmatterHandler(ctx, req, specificArgs)
	case "remove":
		specificArgs := DeleteFrontmatterArgs{
			Path:          args.Path,
			Key:           args.Key,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.RemoveFrontmatterKeyHandler(ctx, req, specificArgs)
	case "add-alias":
		specificArgs := AddAliasArgs{
			Path:          args.Path,
			Alias:         args.Alias,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.AddAliasHandler(ctx, req, specificArgs)
	case "add-tag":
		specificArgs := AddTagArgs{
			Path:          args.Path,
			Tag:           args.Tag,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.AddTagToFrontmatterHandler(ctx, req, specificArgs)
	case "get-inline-fields":
		specificArgs := GetInlineFieldsArgs{
			Path: args.Path,
		}
		return v.GetInlineFieldsHandler(ctx, req, specificArgs)
	case "set-inline-field":
		specificArgs := SetInlineFieldArgs{
			Path:          args.Path,
			Key:           args.Key,
			Value:         args.Value,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.SetInlineFieldHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManageTasksMultiplexArgs multiplexed args
type ManageTasksMultiplexArgs struct {
	Action        string `json:"action" jsonschema:"Action to perform: 'list', 'toggle', 'complete'"`
	Status        string `json:"status,omitempty" jsonschema:"Filter by status: 'all' (default), 'open', 'completed'"`
	Directory     string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum tasks to return (default: all in detailed mode, 100 in compact mode)"`
	Mode          string `json:"mode,omitempty" jsonschema:"Response mode: compact (default) or detailed"`
	Path          string `json:"path,omitempty" jsonschema:"Path to the note"`
	Line          int    `json:"line,omitempty" jsonschema:"Line number of the task (optional if text is provided)"`
	Text          string `json:"text,omitempty" jsonschema:"Text to match the task (partial match, alternative to line number)"`
	ExpectedMtime string `json:"expected_mtime,omitempty" jsonschema:"Expected file modification time (RFC3339Nano) for optimistic concurrency"`
	Texts         string `json:"texts,omitempty" jsonschema:"Comma-separated list or JSON array of task text snippets to mark complete"`
}

// ManageTasksMultiplexHandler routes to the specific handler
func (v *Vault) ManageTasksMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageTasksMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "list":
		specificArgs := ListTasksArgs{
			Status:    args.Status,
			Directory: args.Directory,
			Limit:     args.Limit,
			Mode:      args.Mode,
		}
		return v.ListTasksHandler(ctx, req, specificArgs)
	case "toggle":
		specificArgs := ToggleTaskArgs{
			Path:          args.Path,
			Line:          args.Line,
			Text:          args.Text,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.ToggleTaskHandler(ctx, req, specificArgs)
	case "complete":
		specificArgs := CompleteTasksArgs{
			Path:          args.Path,
			Texts:         args.Texts,
			ExpectedMtime: args.ExpectedMtime,
		}
		return v.CompleteTasksHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// AnalyzeVaultMultiplexArgs multiplexed args
type AnalyzeVaultMultiplexArgs struct {
	Action          string `json:"action" jsonschema:"Action to perform: 'stats', 'broken-links', 'orphan-notes', 'unlinked-mentions', 'find-stubs', 'find-outdated'"`
	Directory       string `json:"directory,omitempty" jsonschema:"Directory to analyze"`
	IncludeDeadEnds bool   `json:"include_no_outgoing,omitempty" jsonschema:"Include notes with no outgoing links (dead ends)"`
	Path            string `json:"path,omitempty" jsonschema:"Path to the note to find unlinked mentions of"`
	MaxWords        int    `json:"max_words,omitempty" jsonschema:"Maximum word count to qualify as stub (default 100)"`
	Limit           int    `json:"limit,omitempty" jsonschema:"Maximum results (default 50)"`
	Days            int    `json:"days,omitempty" jsonschema:"Days since modification to qualify as outdated (default 90)"`
}

// AnalyzeVaultMultiplexHandler routes to the specific handler
func (v *Vault) AnalyzeVaultMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args AnalyzeVaultMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "stats":
		specificArgs := VaultStatsArgs{
			Directory: args.Directory,
		}
		return v.VaultStatsHandler(ctx, req, specificArgs)
	case "broken-links":
		specificArgs := BrokenLinksArgs{
			Directory: args.Directory,
		}
		return v.BrokenLinksHandler(ctx, req, specificArgs)
	case "orphan-notes":
		specificArgs := OrphanNotesArgs{
			Directory:       args.Directory,
			IncludeDeadEnds: args.IncludeDeadEnds,
		}
		return v.OrphanNotesHandler(ctx, req, specificArgs)
	case "unlinked-mentions":
		specificArgs := UnlinkedMentionsArgs{
			Path: args.Path,
		}
		return v.UnlinkedMentionsHandler(ctx, req, specificArgs)
	case "find-stubs":
		specificArgs := FindStubsArgs{
			MaxWords:  args.MaxWords,
			Directory: args.Directory,
			Limit:     args.Limit,
		}
		return v.FindStubsHandler(ctx, req, specificArgs)
	case "find-outdated":
		specificArgs := FindOutdatedArgs{
			Days:      args.Days,
			Directory: args.Directory,
			Limit:     args.Limit,
		}
		return v.FindOutdatedHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManageCanvasMultiplexArgs multiplexed args
type ManageCanvasMultiplexArgs struct {
	Action       string `json:"action" jsonschema:"Action to perform: 'list', 'read', 'create', 'add-node', 'add-edge'"`
	Directory    string `json:"directory,omitempty" jsonschema:"Root directory to list from"`
	IncludeEmpty bool   `json:"include_empty,omitempty" jsonschema:"Whether to include empty directories (default true)"`
	Path         string `json:"path,omitempty" jsonschema:"Path to the note relative to vault root"`
	Content      string `json:"content,omitempty" jsonschema:"Initial content (JSON)"`
	Canvas       string `json:"canvas,omitempty" jsonschema:"Path to canvas file"`
	Type         string `json:"type,omitempty" jsonschema:"Node type: 'text' (default), 'file', 'link', 'group'"`
	X            int    `json:"x,omitempty" jsonschema:"X position"`
	Y            int    `json:"y,omitempty" jsonschema:"Y position"`
	Width        int    `json:"width,omitempty" jsonschema:"Node width"`
	Height       int    `json:"height,omitempty" jsonschema:"Node height"`
	Label        string `json:"label,omitempty" jsonschema:"Node label (optional)"`
	From         string `json:"from,omitempty" jsonschema:"Source node ID"`
	To           string `json:"to,omitempty" jsonschema:"Target node ID"`
}

// ManageCanvasMultiplexHandler routes to the specific handler
func (v *Vault) ManageCanvasMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageCanvasMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "list":
		specificArgs := ListDirsArgs{
			Directory:    args.Directory,
			IncludeEmpty: args.IncludeEmpty,
		}
		return v.ListCanvasesHandler(ctx, req, specificArgs)
	case "read":
		specificArgs := ReadNoteArgs{
			Path: args.Path,
		}
		return v.ReadCanvasHandler(ctx, req, specificArgs)
	case "create":
		specificArgs := CreateCanvasArgs{
			Path:    args.Path,
			Content: args.Content,
		}
		return v.CreateCanvasHandler(ctx, req, specificArgs)
	case "add-node":
		specificArgs := AddNodeArgs{
			Canvas:  args.Canvas,
			Type:    args.Type,
			Content: args.Content,
			X:       args.X,
			Y:       args.Y,
			Width:   args.Width,
			Height:  args.Height,
			Label:   args.Label,
		}
		return v.AddCanvasNodeHandler(ctx, req, specificArgs)
	case "add-edge":
		specificArgs := ConnectNodesArgs{
			Canvas: args.Canvas,
			From:   args.From,
			To:     args.To,
			Label:  args.Label,
		}
		return v.AddCanvasEdgeHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManageMocsMultiplexArgs multiplexed args
type ManageMocsMultiplexArgs struct {
	Action         string `json:"action" jsonschema:"Action to perform: 'discover', 'generate', 'update', 'generate-index'"`
	Directory      string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	Title          string `json:"title,omitempty" jsonschema:"MOC title"`
	Output         string `json:"output,omitempty" jsonschema:"Output file path"`
	GroupBy        string `json:"group_by,omitempty" jsonschema:"Group by: 'none' (default), 'tag', 'alpha'"`
	Recursive      bool   `json:"recursive,omitempty" jsonschema:"Include subdirectories"`
	Path           string `json:"path,omitempty" jsonschema:"Path to the MOC file"`
	IncludeOrphans bool   `json:"include_orphans,omitempty" jsonschema:"Include notes without links/tags (default true)"`
}

// ManageMocsMultiplexHandler routes to the specific handler
func (v *Vault) ManageMocsMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageMocsMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "discover":
		specificArgs := DiscoverMOCsArgs{
			Directory: args.Directory,
		}
		return v.DiscoverMOCsHandler(ctx, req, specificArgs)
	case "generate":
		specificArgs := GenerateMOCArgs{
			Directory: args.Directory,
			Title:     args.Title,
			Output:    args.Output,
			GroupBy:   args.GroupBy,
			Recursive: args.Recursive,
		}
		return v.GenerateMOCHandler(ctx, req, specificArgs)
	case "update":
		specificArgs := UpdateMOCArgs{
			Path:      args.Path,
			Directory: args.Directory,
			Recursive: args.Recursive,
		}
		return v.UpdateMOCHandler(ctx, req, specificArgs)
	case "generate-index":
		specificArgs := GenerateIndexArgs{
			Directory:      args.Directory,
			Output:         args.Output,
			Title:          args.Title,
			IncludeOrphans: args.IncludeOrphans,
		}
		return v.GenerateIndexHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ReadBatchMultiplexArgs multiplexed args
type ReadBatchMultiplexArgs struct {
	Action             string `json:"action" jsonschema:"Action to perform: 'read', 'get-section', 'get-headings', 'get-summary'"`
	Paths              string `json:"paths,omitempty" jsonschema:"Comma-separated list or JSON array of paths"`
	IncludeFrontmatter bool   `json:"include_frontmatter,omitempty" jsonschema:"Include frontmatter in output (default true)"`
	Path               string `json:"path,omitempty" jsonschema:"Path to the note"`
	Heading            string `json:"heading,omitempty" jsonschema:"Heading to extract"`
	Lines              int    `json:"lines,omitempty" jsonschema:"Number of preview lines (default 5)"`
}

// ReadBatchMultiplexHandler routes to the specific handler
func (v *Vault) ReadBatchMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ReadBatchMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "read":
		specificArgs := ReadNotesArgs{
			Paths:              args.Paths,
			IncludeFrontmatter: args.IncludeFrontmatter,
		}
		return v.ReadNotesHandler(ctx, req, specificArgs)
	case "get-section":
		specificArgs := GetSectionArgs{
			Path:    args.Path,
			Heading: args.Heading,
		}
		return v.GetSectionHandler(ctx, req, specificArgs)
	case "get-headings":
		specificArgs := GetHeadingsArgs{
			Path: args.Path,
		}
		return v.GetHeadingsHandler(ctx, req, specificArgs)
	case "get-summary":
		specificArgs := GetNoteSummaryArgs{
			Path:  args.Path,
			Lines: args.Lines,
		}
		return v.GetNoteSummaryHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManageLinksMultiplexArgs multiplexed args
type ManageLinksMultiplexArgs struct {
	Action string `json:"action" jsonschema:"Action to perform: 'backlinks', 'forward-links', 'suggest'"`
	Path   string `json:"path,omitempty" jsonschema:"Path to the note"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum results (default 10)"`
}

// ManageLinksMultiplexHandler routes to the specific handler
func (v *Vault) ManageLinksMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageLinksMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "backlinks":
		specificArgs := GetBacklinksArgs{
			Path: args.Path,
		}
		return v.BacklinksHandler(ctx, req, specificArgs)
	case "forward-links":
		specificArgs := ForwardLinksArgs{
			Path: args.Path,
		}
		return v.ForwardLinksHandler(ctx, req, specificArgs)
	case "suggest":
		specificArgs := SuggestLinksArgs{
			Path:  args.Path,
			Limit: args.Limit,
		}
		return v.SuggestLinksHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// BulkOperationsMultiplexArgs multiplexed args
type BulkOperationsMultiplexArgs struct {
	Action      string `json:"action" jsonschema:"Action to perform: 'tag', 'move', 'set-frontmatter'"`
	Paths       string `json:"paths,omitempty" jsonschema:"Comma-separated list or JSON array of paths"`
	Tag         string `json:"tag,omitempty" jsonschema:"Tag to add or remove"`
	TagAction   string `json:"tag_action,omitempty" jsonschema:"Action: 'add' (default), 'remove'"`
	DryRun      bool   `json:"dry_run,omitempty" jsonschema:"Preview changes without modifying files"`
	Destination string `json:"destination,omitempty" jsonschema:"Destination folder"`
	UpdateLinks bool   `json:"update_links,omitempty" jsonschema:"Whether to update links (default true)"`
	Key         string `json:"key,omitempty" jsonschema:"Frontmatter key"`
	Value       string `json:"value,omitempty" jsonschema:"Value to set"`
}

// BulkOperationsMultiplexHandler routes to the specific handler
func (v *Vault) BulkOperationsMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args BulkOperationsMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "tag":
		specificArgs := BulkTagArgs{
			Paths:  args.Paths,
			Tag:    args.Tag,
			Action: args.Action,
			DryRun: args.DryRun,
		}
		return v.BulkTagHandler(ctx, req, specificArgs)
	case "move":
		specificArgs := BulkMoveArgs{
			Paths:       args.Paths,
			Destination: args.Destination,
			UpdateLinks: args.UpdateLinks,
			DryRun:      args.DryRun,
		}
		return v.BulkMoveHandler(ctx, req, specificArgs)
	case "set-frontmatter":
		specificArgs := BulkSetFrontmatterArgs{
			Paths:  args.Paths,
			Key:    args.Key,
			Value:  args.Value,
			DryRun: args.DryRun,
		}
		return v.BulkSetFrontmatterHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManageTemplatesMultiplexArgs multiplexed args
type ManageTemplatesMultiplexArgs struct {
	Action         string `json:"action" jsonschema:"Action to perform: 'list', 'get', 'apply'"`
	Folder         string `json:"folder,omitempty" jsonschema:"Templates folder (default: 'templates')"`
	Name           string `json:"name,omitempty" jsonschema:"Template name"`
	Template       string `json:"template,omitempty" jsonschema:"Template name"`
	Path           string `json:"path,omitempty" jsonschema:"Target note path"`
	TemplateFolder string `json:"template_folder,omitempty" jsonschema:"Templates folder (default: 'templates')"`
	Variables      string `json:"variables,omitempty" jsonschema:"JSON string or key=value pairs of variables"`
}

// ManageTemplatesMultiplexHandler routes to the specific handler
func (v *Vault) ManageTemplatesMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageTemplatesMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "list":
		specificArgs := ListTemplatesArgs{
			Folder: args.Folder,
		}
		return v.ListTemplatesHandler(ctx, req, specificArgs)
	case "get":
		specificArgs := GetTemplateArgs{
			Name:   args.Name,
			Folder: args.Folder,
		}
		return v.GetTemplateHandler(ctx, req, specificArgs)
	case "apply":
		specificArgs := ApplyTemplateArgs{
			Template:       args.Template,
			Path:           args.Path,
			TemplateFolder: args.TemplateFolder,
			Variables:      args.Variables,
		}
		return v.ApplyTemplateHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// RefactorNotesMultiplexArgs multiplexed args
type RefactorNotesMultiplexArgs struct {
	Action             string `json:"action" jsonschema:"Action to perform: 'split', 'merge', 'extract-section'"`
	Path               string `json:"path,omitempty" jsonschema:"Source note path"`
	Level              int    `json:"level,omitempty" jsonschema:"Heading level to split at (default: 2, for split action)"`
	KeepOriginal       bool   `json:"keep_original,omitempty" jsonschema:"Keep extracted content in original note (for split action)"`
	OutputDir          string `json:"output_dir,omitempty" jsonschema:"Directory for new notes (for split action)"`
	DryRun             bool   `json:"dry_run,omitempty" jsonschema:"Preview changes without modifying files"`
	Paths              string `json:"paths,omitempty" jsonschema:"Comma-separated list of notes to merge (for merge action)"`
	Output             string `json:"output,omitempty" jsonschema:"Output note path"`
	Separator          string `json:"separator,omitempty" jsonschema:"Separator between notes (for merge action)"`
	DeleteOriginals    bool   `json:"delete_originals,omitempty" jsonschema:"Delete original notes after merge (for merge action)"`
	AddHeadings        bool   `json:"add_headings,omitempty" jsonschema:"Add note names as headings (for merge action)"`
	Heading            string `json:"heading,omitempty" jsonschema:"Heading to extract (for extract-section action)"`
	RemoveFromOriginal bool   `json:"remove_from_original,omitempty" jsonschema:"Remove from source note (default true, for extract-section action)"`
	AddLink            bool   `json:"add_link,omitempty" jsonschema:"Add link to new note in source (default true, for extract-section action)"`
}

// RefactorNotesMultiplexHandler routes to the specific handler
func (v *Vault) RefactorNotesMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args RefactorNotesMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "split":
		specificArgs := ExtractNoteArgs{
			Path:         args.Path,
			Level:        args.Level,
			KeepOriginal: args.KeepOriginal,
			OutputDir:    args.OutputDir,
			DryRun:       args.DryRun,
		}
		return v.ExtractNoteHandler(ctx, req, specificArgs)
	case "merge":
		specificArgs := MergeNotesArgs{
			Paths:           args.Paths,
			Output:          args.Output,
			Separator:       args.Separator,
			DeleteOriginals: args.DeleteOriginals,
			AddHeadings:     args.AddHeadings,
			DryRun:          args.DryRun,
		}
		return v.MergeNotesHandler(ctx, req, specificArgs)
	case "extract-section":
		specificArgs := ExtractSectionArgs{
			Path:               args.Path,
			Heading:            args.Heading,
			Output:             args.Output,
			RemoveFromOriginal: args.RemoveFromOriginal,
			AddLink:            args.AddLink,
			DryRun:             args.DryRun,
		}
		return v.ExtractSectionHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}

// ManageVaultsMultiplexArgs multiplexed args
type ManageVaultsMultiplexArgs struct {
	Action string `json:"action" jsonschema:"Action to perform: 'list', 'switch'"`
	Vault  string `json:"vault,omitempty" jsonschema:"The alias or absolute path of the vault to switch to"`
}

// ManageVaultsMultiplexHandler routes to the specific handler
func (v *Vault) ManageVaultsMultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args ManageVaultsMultiplexArgs) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "list":
		specificArgs := ListVaultsArgs{}
		return v.ListVaultsHandler(ctx, req, specificArgs)
	case "switch":
		specificArgs := SwitchVaultArgs{
			Vault: args.Vault,
		}
		return v.SwitchVaultHandler(ctx, req, specificArgs)
	default:
		return nil, nil, fmt.Errorf("unknown action: %s", args.Action)
	}
}
