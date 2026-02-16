package vault

// --- Common ---

// ListNotesArgs arguments for list-notes
type ListNotesArgs struct {
	Directory string `json:"directory,omitempty" jsonschema:"Directory path relative to vault root (optional)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of notes to return (optional, 0 = no limit)"`
	Offset    int    `json:"offset,omitempty" jsonschema:"Number of notes to skip for pagination (optional, default 0)"`
}

// WriteNoteArgs arguments for write-note
type WriteNoteArgs struct {
	Path    string `json:"path" jsonschema:"Path to the note relative to vault root (.md extension required)"`
	Content string `json:"content" jsonschema:"Content of the note"`
}

// ReadNoteArgs arguments for read-note
type ReadNoteArgs struct {
	Path string `json:"path" jsonschema:"Path to the note relative to vault root"`
}

// DeleteNoteArgs arguments for delete-note
type DeleteNoteArgs struct {
	Path string `json:"path" jsonschema:"Path to the note to delete"`
}

// ReadNotesArgs arguments for read-multiple-notes
type ReadNotesArgs struct {
	Paths              string `json:"paths" jsonschema:"Comma-separated list or JSON array of paths"`
	IncludeFrontmatter bool   `json:"include_frontmatter,omitempty" jsonschema:"Include frontmatter in output (default true)"`
}

// GetNoteSummaryArgs arguments for get-note-summary
type GetNoteSummaryArgs struct {
	Path  string `json:"path" jsonschema:"Path to the note"`
	Lines int    `json:"lines,omitempty" jsonschema:"Number of preview lines (default 5)"`
}

// GetSectionArgs arguments for get-section
type GetSectionArgs struct {
	Path    string `json:"path" jsonschema:"Path to the note"`
	Heading string `json:"heading" jsonschema:"Heading to extract"`
}

// GetHeadingsArgs arguments for get-headings
type GetHeadingsArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
}

// SearchHeadingsArgs arguments for search-headings
type SearchHeadingsArgs struct {
	Query     string `json:"query" jsonschema:"Search query"`
	Level     int    `json:"level,omitempty" jsonschema:"Heading level to filter (0 for all)"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
}

// --- Editing ---

// EditNoteArgs arguments for edit-note
type EditNoteArgs struct {
	Path         string `json:"path" jsonschema:"Path to the note"`
	OldText      string `json:"old_text" jsonschema:"Text to find and replace"`
	NewText      string `json:"new_text" jsonschema:"Replacement text"`
	ReplaceAll   bool   `json:"replace_all,omitempty" jsonschema:"Whether to replace all occurrences (default false)"`
	ContextLines int    `json:"context_lines,omitempty" jsonschema:"Number of context lines to return (default 0)"`
}

// ReplaceSectionArgs arguments for replace-section
type ReplaceSectionArgs struct {
	Path         string `json:"path" jsonschema:"Path to the note"`
	Heading      string `json:"heading" jsonschema:"Heading of the section to replace"`
	Content      string `json:"content" jsonschema:"New content for the section"`
	ContextLines int    `json:"context_lines,omitempty" jsonschema:"Number of context lines to return (default 0)"`
}

// EditEntry represents a single edit in a batch
type EditEntry struct {
	OldText string `json:"old_text" jsonschema:"Text to find"`
	NewText string `json:"new_text" jsonschema:"Replacement text"`
}

// BatchEditArgs arguments for batch-edit-note
type BatchEditArgs struct {
	Path         string      `json:"path" jsonschema:"Path to the note"`
	Edits        []EditEntry `json:"edits" jsonschema:"List of edits to apply"`
	ContextLines int         `json:"context_lines,omitempty" jsonschema:"Number of context lines to return (default 0)"`
}

// AppendNoteArgs arguments for append-note
type AppendNoteArgs struct {
	Path         string `json:"path" jsonschema:"Path to the note"`
	Content      string `json:"content" jsonschema:"Content to append"`
	Position     string `json:"position,omitempty" jsonschema:"Position to insert: 'end' (default), 'start', 'before', 'after'"`
	After        string `json:"after,omitempty" jsonschema:"Heading or text to insert after (if position is 'after')"`
	Before       string `json:"before,omitempty" jsonschema:"Heading or text to insert before (if position is 'before')"`
	ContextLines int    `json:"context_lines,omitempty" jsonschema:"Number of context lines to return (default 0)"`
}

// --- Search ---

// SearchArgs arguments for search-vault
type SearchArgs struct {
	Query     string `json:"query" jsonschema:"Search query"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
}

// SearchAdvancedArgs arguments for search-advanced
type SearchAdvancedArgs struct {
	Query     string `json:"query" jsonschema:"Search query"`
	SearchIn  string `json:"in,omitempty" jsonschema:"Where to search: 'content' (default), 'file', 'heading', 'block'"`
	Operator  string `json:"operator,omitempty" jsonschema:"Logical operator: 'and' (default), 'or'"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum results to return (default 50)"`
}

// SearchDateArgs arguments for search-date
type SearchDateArgs struct {
	From      string `json:"from,omitempty" jsonschema:"Start date (YYYY-MM-DD)"`
	To        string `json:"to,omitempty" jsonschema:"End date (YYYY-MM-DD)"`
	DateType  string `json:"type,omitempty" jsonschema:"Date type to check: 'modified' (default), 'created'"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum results to return (default 50)"`
}

// SearchRegexArgs arguments for search-regex
type SearchRegexArgs struct {
	Pattern         string `json:"pattern" jsonschema:"Regex pattern"`
	Directory       string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	Limit           int    `json:"limit,omitempty" jsonschema:"Maximum results to return (default 50)"`
	CaseInsensitive bool   `json:"case_insensitive,omitempty" jsonschema:"Whether to ignore case (default true)"`
}

// --- Tasks ---

// ListTasksArgs arguments for list-tasks
type ListTasksArgs struct {
	Status    string `json:"status,omitempty" jsonschema:"Filter by status: 'all' (default), 'open', 'completed'"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
}

// ToggleTaskArgs arguments for toggle-task
type ToggleTaskArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
	Line int    `json:"line" jsonschema:"Line number of the task"`
}

// --- Tags ---

// SearchTagsArgs arguments for search-by-tags
type SearchTagsArgs struct {
	Tags      string `json:"tags" jsonschema:"Comma-separated list of tags"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
}

// --- Bulk ---

// BulkTagArgs arguments for bulk-tag
type BulkTagArgs struct {
	Paths  string `json:"paths" jsonschema:"Comma-separated list or JSON array of paths"`
	Tag    string `json:"tag" jsonschema:"Tag to add or remove"`
	Action string `json:"action,omitempty" jsonschema:"Action: 'add' (default), 'remove'"`
}

// BulkMoveArgs arguments for bulk-move
type BulkMoveArgs struct {
	Paths       string `json:"paths" jsonschema:"Comma-separated list or JSON array of paths"`
	Destination string `json:"destination" jsonschema:"Destination folder"`
	UpdateLinks bool   `json:"update_links,omitempty" jsonschema:"Whether to update links (default true)"`
}

// BulkSetFrontmatterArgs arguments for bulk-set-frontmatter
type BulkSetFrontmatterArgs struct {
	Paths string `json:"paths" jsonschema:"Comma-separated list or JSON array of paths"`
	Key   string `json:"key" jsonschema:"Frontmatter key"`
	Value string `json:"value" jsonschema:"Value to set"`
}

// --- Folders ---

// ListDirsArgs arguments for list-directories
type ListDirsArgs struct {
	Directory    string `json:"directory,omitempty" jsonschema:"Root directory to list from"`
	IncludeEmpty bool   `json:"include_empty,omitempty" jsonschema:"Whether to include empty directories (default true)"`
}

// CreateDirArgs arguments for create-directory
type CreateDirArgs struct {
	Path string `json:"path" jsonschema:"Path of the directory to create"`
}

// MoveArgs arguments for move-file
type MoveArgs struct {
	Source      string `json:"source" jsonschema:"Source path"`
	Destination string `json:"destination" jsonschema:"Destination path"`
	UpdateLinks bool   `json:"update_links,omitempty" jsonschema:"Whether to update links to this file (default true)"`
}

// DeleteDirArgs arguments for delete-directory
type DeleteDirArgs struct {
	Path  string `json:"path" jsonschema:"Path of the directory to delete"`
	Force bool   `json:"force,omitempty" jsonschema:"Force delete even if not empty (default false)"`
}

// --- Periodic Notes ---

// DailyNoteArgs arguments for daily-note
type DailyNoteArgs struct {
	Date            string `json:"date,omitempty" jsonschema:"Date string (default: today)"`
	Folder          string `json:"folder,omitempty" jsonschema:"Folder for daily notes (default: 'daily')"`
	Format          string `json:"format,omitempty" jsonschema:"Date format (default: '2006-01-02')"`
	CreateIfMissing bool   `json:"create,omitempty" jsonschema:"Create if missing (default: true)"`
}

// WeeklyNoteArgs arguments for weekly-note
type WeeklyNoteArgs struct {
	Date            string `json:"date,omitempty" jsonschema:"Date string (default: today)"`
	Folder          string `json:"folder,omitempty" jsonschema:"Folder for weekly notes (default: 'weekly')"`
	Format          string `json:"format,omitempty" jsonschema:"Date format (default: '2006-W02')"`
	CreateIfMissing bool   `json:"create,omitempty" jsonschema:"Create if missing (default: true)"`
}

// MonthlyNoteArgs arguments for monthly-note
type MonthlyNoteArgs struct {
	Date            string `json:"date,omitempty" jsonschema:"Date string (default: today)"`
	Folder          string `json:"folder,omitempty" jsonschema:"Folder for monthly notes (default: 'monthly')"`
	Format          string `json:"format,omitempty" jsonschema:"Date format (default: '2006-01')"`
	CreateIfMissing bool   `json:"create,omitempty" jsonschema:"Create if missing (default: true)"`
}

// QuarterlyNoteArgs arguments for quarterly-note
type QuarterlyNoteArgs struct {
	Date            string `json:"date,omitempty" jsonschema:"Date string (default: today)"`
	Folder          string `json:"folder,omitempty" jsonschema:"Folder for quarterly notes (default: 'quarterly')"`
	CreateIfMissing bool   `json:"create,omitempty" jsonschema:"Create if missing (default: true)"`
}

// YearlyNoteArgs arguments for yearly-note
type YearlyNoteArgs struct {
	Date            string `json:"date,omitempty" jsonschema:"Date string (default: today)"`
	Folder          string `json:"folder,omitempty" jsonschema:"Folder for yearly notes (default: 'yearly')"`
	CreateIfMissing bool   `json:"create,omitempty" jsonschema:"Create if missing (default: true)"`
}

// ListPeriodicArgs arguments for list-periodic
type ListPeriodicArgs struct {
	Type   string `json:"type,omitempty" jsonschema:"Type of note: 'daily', 'weekly', 'monthly', 'quarterly', 'yearly'"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of notes to return"`
	Folder string `json:"folder,omitempty" jsonschema:"Folder to search in"`
}

// --- Templates ---

// ListTemplatesArgs arguments for list-templates
type ListTemplatesArgs struct {
	Folder string `json:"folder,omitempty" jsonschema:"Templates folder (default: 'templates')"`
}

// GetTemplateArgs arguments for get-template
type GetTemplateArgs struct {
	Name   string `json:"name" jsonschema:"Template name"`
	Folder string `json:"folder,omitempty" jsonschema:"Templates folder (default: 'templates')"`
}

// ApplyTemplateArgs arguments for apply-template
type ApplyTemplateArgs struct {
	Template       string `json:"template" jsonschema:"Template name"`
	Path           string `json:"path" jsonschema:"Target note path"`
	TemplateFolder string `json:"template_folder,omitempty" jsonschema:"Templates folder (default: 'templates')"`
	Variables      string `json:"variables,omitempty" jsonschema:"JSON string or key=value pairs of variables"`
}

// --- Frontmatter ---

// QueryFrontmatterArgs arguments for query-frontmatter
type QueryFrontmatterArgs struct {
	Query     string `json:"query" jsonschema:"Query string (key=value or key:value)"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
}

// GetFrontmatterArgs arguments for get-frontmatter
type GetFrontmatterArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
}

// SetFrontmatterArgs arguments for set-frontmatter
type SetFrontmatterArgs struct {
	Path  string `json:"path" jsonschema:"Path to the note"`
	Key   string `json:"key" jsonschema:"Frontmatter key"`
	Value string `json:"value" jsonschema:"Value to set"`
}

// DeleteFrontmatterArgs arguments for delete-frontmatter
type DeleteFrontmatterArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
	Key  string `json:"key" jsonschema:"Frontmatter key to remove"`
}

// AddAliasArgs arguments for add-alias
type AddAliasArgs struct {
	Path  string `json:"path" jsonschema:"Path to the note"`
	Alias string `json:"alias" jsonschema:"Alias to add"`
}

// AddTagArgs arguments for add-tag
type AddTagArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
	Tag  string `json:"tag" jsonschema:"Tag to add"`
}

// RemoveTagArgs arguments for remove-tag
type RemoveTagArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
	Tag  string `json:"tag" jsonschema:"Tag to remove"`
}

// --- Graph ---

// ForwardLinksArgs arguments for forward-links
type ForwardLinksArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
}

// OrphanNotesArgs arguments for orphan-notes
type OrphanNotesArgs struct {
	Directory       string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	IncludeDeadEnds bool   `json:"include_no_outgoing,omitempty" jsonschema:"Include notes with no outgoing links (dead ends)"`
}

// BrokenLinksArgs arguments for broken-links
type BrokenLinksArgs struct {
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
}

// GraphNeighborsArgs arguments for graph-neighbors
type GraphNeighborsArgs struct {
	Path            string `json:"path" jsonschema:"Path to the note"`
	Directory       string `json:"directory,omitempty" jsonschema:"Directory scope"`
	IncludeDeadEnds bool   `json:"include_no_outgoing,omitempty" jsonschema:"Include nodes with no outgoing links"`
}

// --- Inline Fields ---

// GetInlineFieldsArgs arguments for get-inline-fields
type GetInlineFieldsArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
}

// SetInlineFieldArgs arguments for set-inline-field
type SetInlineFieldArgs struct {
	Path  string `json:"path" jsonschema:"Path to the note"`
	Key   string `json:"key" jsonschema:"Field key"`
	Value string `json:"value" jsonschema:"Field value"`
}

// SearchInlineFieldsArgs arguments for search-inline-fields
type SearchInlineFieldsArgs struct {
	Key       string `json:"key" jsonschema:"Field key"`
	Value     string `json:"value,omitempty" jsonschema:"Field value to match (optional)"`
	Operator  string `json:"operator,omitempty" jsonschema:"Operator: 'contains' (default), 'equals', 'gt', 'lt'"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to search in"`
}

// --- Links ---

// GetBacklinksArgs arguments for get-backlinks
type GetBacklinksArgs struct {
	Path string `json:"path" jsonschema:"Path to the note"`
}

// RenameNoteArgs arguments for rename-note
type RenameNoteArgs struct {
	OldPath string `json:"old_path" jsonschema:"Old note path"`
	NewPath string `json:"new_path" jsonschema:"New note path"`
}

// UpdateLinksArgs arguments for update-links
type UpdateLinksArgs struct {
	OldPath string `json:"old_path" jsonschema:"Old note path"`
	NewPath string `json:"new_path" jsonschema:"New note path"`
}

// --- MOC ---

// DiscoverMOCsArgs arguments for discover-mocs
type DiscoverMOCsArgs struct {
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
}

// GenerateMOCArgs arguments for generate-moc
type GenerateMOCArgs struct {
	Directory string `json:"directory,omitempty" jsonschema:"Directory to generate MOC for"`
	Title     string `json:"title,omitempty" jsonschema:"MOC title"`
	Output    string `json:"output,omitempty" jsonschema:"Output file path"`
	GroupBy   string `json:"group_by,omitempty" jsonschema:"Group by: 'none' (default), 'tag', 'alpha'"`
	Recursive bool   `json:"recursive,omitempty" jsonschema:"Include subdirectories"`
}

// UpdateMOCArgs arguments for update-moc
type UpdateMOCArgs struct {
	Path      string `json:"path" jsonschema:"Path to the MOC file"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to scan"`
	Recursive bool   `json:"recursive,omitempty" jsonschema:"Include subdirectories"`
}

// GenerateIndexArgs arguments for generate-index
type GenerateIndexArgs struct {
	Directory      string `json:"directory,omitempty" jsonschema:"Directory to index"`
	Output         string `json:"output,omitempty" jsonschema:"Output file path"`
	Title          string `json:"title,omitempty" jsonschema:"Index title (default: 'Index')"`
	IncludeOrphans bool   `json:"include_orphans,omitempty" jsonschema:"Include notes without links/tags (default true)"`
}

// --- Refactor ---

// ExtractNoteArgs arguments for extract-note
type ExtractNoteArgs struct {
	Path         string `json:"path" jsonschema:"Source note path"`
	Level        int    `json:"level,omitempty" jsonschema:"Heading level to extract (default: 2)"`
	KeepOriginal bool   `json:"keep_original,omitempty" jsonschema:"Keep extracted content in original note"`
	OutputDir    string `json:"output_dir,omitempty" jsonschema:"Directory for new notes"`
}

// MergeNotesArgs arguments for merge-notes
type MergeNotesArgs struct {
	Paths           string `json:"paths" jsonschema:"Comma-separated list of notes to merge"`
	Output          string `json:"output" jsonschema:"Output note path"`
	Separator       string `json:"separator,omitempty" jsonschema:"Separator between notes"`
	DeleteOriginals bool   `json:"delete_originals,omitempty" jsonschema:"Delete original notes after merge"`
	AddHeadings     bool   `json:"add_headings,omitempty" jsonschema:"Add note names as headings"`
}

// ExtractSectionArgs arguments for extract-section
type ExtractSectionArgs struct {
	Path               string `json:"path" jsonschema:"Source note path"`
	Heading            string `json:"heading" jsonschema:"Heading to extract"`
	Output             string `json:"output" jsonschema:"Output note path"`
	RemoveFromOriginal bool   `json:"remove_from_original,omitempty" jsonschema:"Remove from source note (default true)"`
	AddLink            bool   `json:"add_link,omitempty" jsonschema:"Add link to new note in source (default true)"`
}

// DuplicateNoteArgs arguments for duplicate-note
type DuplicateNoteArgs struct {
	Path   string `json:"path" jsonschema:"Source note path"`
	Output string `json:"output,omitempty" jsonschema:"Output note path"`
}

// --- Stats ---

// VaultStatsArgs arguments for vault-stats
type VaultStatsArgs struct {
	Directory string `json:"directory,omitempty" jsonschema:"Directory to analyze"`
}

// --- Analysis ---

// FindStubsArgs arguments for find-stubs
type FindStubsArgs struct {
	MaxWords  int    `json:"max_words,omitempty" jsonschema:"Maximum word count to qualify as stub (default 100)"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum results (default 50)"`
}

// FindOutdatedArgs arguments for find-outdated
type FindOutdatedArgs struct {
	Days      int    `json:"days,omitempty" jsonschema:"Days since modification to qualify as outdated (default 90)"`
	Directory string `json:"directory,omitempty" jsonschema:"Directory to limit search to"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum results (default 50)"`
}

// UnlinkedMentionsArgs arguments for unlinked-mentions
type UnlinkedMentionsArgs struct {
	Path string `json:"path" jsonschema:"Path to the note to find unlinked mentions of"`
}

// SuggestLinksArgs arguments for suggest-links
type SuggestLinksArgs struct {
	Path  string `json:"path" jsonschema:"Path to the note"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum results (default 10)"`
}

// --- Canvas ---

// CreateCanvasArgs arguments for create-canvas
type CreateCanvasArgs struct {
	Path    string `json:"path" jsonschema:"Path for the new canvas"`
	Content string `json:"content,omitempty" jsonschema:"Initial content (JSON)"`
}

// AddNodeArgs arguments for canvas-add-node
type AddNodeArgs struct {
	Canvas  string `json:"canvas" jsonschema:"Path to canvas file"`
	Type    string `json:"type,omitempty" jsonschema:"Node type: 'text' (default), 'file', 'link', 'group'"`
	Content string `json:"content" jsonschema:"Content of the node"`
	X       int    `json:"x,omitempty" jsonschema:"X position"`
	Y       int    `json:"y,omitempty" jsonschema:"Y position"`
	Width   int    `json:"width,omitempty" jsonschema:"Node width"`
	Height  int    `json:"height,omitempty" jsonschema:"Node height"`
	Label   string `json:"label,omitempty" jsonschema:"Node label (optional)"`
}

// ConnectNodesArgs arguments for canvas-connect-nodes
type ConnectNodesArgs struct {
	Canvas string `json:"canvas" jsonschema:"Path to canvas file"`
	From   string `json:"from" jsonschema:"Source node ID"`
	To     string `json:"to" jsonschema:"Target node ID"`
	Label  string `json:"label,omitempty" jsonschema:"Edge label"`
}
