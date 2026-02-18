package vault

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListHandlersRejectPathTraversal(t *testing.T) {
	ctx := context.Background()
	v, _ := setupTestVault(t)

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "list-notes",
			run: func() error {
				_, _, err := v.ListNotesHandler(ctx, nil, ListNotesArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "list-folders",
			run: func() error {
				_, _, err := v.ListFoldersHandler(ctx, nil, ListDirsArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "list-daily-notes",
			run: func() error {
				_, _, err := v.ListDailyNotesHandler(ctx, nil, ListPeriodicArgs{Folder: "../"})
				return err
			},
		},
		{
			name: "list-templates",
			run: func() error {
				_, _, err := v.ListTemplatesHandler(ctx, nil, ListTemplatesArgs{Folder: "../"})
				return err
			},
		},
		{
			name: "query-frontmatter",
			run: func() error {
				_, _, err := v.QueryFrontmatterHandler(ctx, nil, QueryFrontmatterArgs{
					Query:     "status=active",
					Directory: "../",
				})
				return err
			},
		},
		{
			name: "search-by-tags",
			run: func() error {
				_, _, err := v.SearchByTagsHandler(ctx, nil, SearchTagsArgs{
					Tags:      "project",
					Directory: "../",
				})
				return err
			},
		},
		{
			name: "list-tasks",
			run: func() error {
				_, _, err := v.ListTasksHandler(ctx, nil, ListTasksArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "find-stubs",
			run: func() error {
				_, _, err := v.FindStubsHandler(ctx, nil, FindStubsArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "find-outdated",
			run: func() error {
				_, _, err := v.FindOutdatedHandler(ctx, nil, FindOutdatedArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "vault-stats",
			run: func() error {
				_, _, err := v.VaultStatsHandler(ctx, nil, VaultStatsArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "orphan-notes",
			run: func() error {
				_, _, err := v.OrphanNotesHandler(ctx, nil, OrphanNotesArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "broken-links",
			run: func() error {
				_, _, err := v.BrokenLinksHandler(ctx, nil, BrokenLinksArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "list-canvases",
			run: func() error {
				_, _, err := v.ListCanvasesHandler(ctx, nil, ListDirsArgs{Directory: "../"})
				return err
			},
		},
		{
			name: "search-headings",
			run: func() error {
				_, _, err := v.SearchHeadingsHandler(ctx, nil, SearchHeadingsArgs{
					Query:     "heading",
					Directory: "../",
				})
				return err
			},
		},
		{
			name: "search-inline-fields",
			run: func() error {
				_, _, err := v.QueryInlineFieldsHandler(ctx, nil, SearchInlineFieldsArgs{
					Key:       "status",
					Directory: "../",
				})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil {
				t.Fatal("expected traversal rejection error")
			}
			if !strings.Contains(err.Error(), "within vault") {
				t.Fatalf("expected vault-bound error, got: %v", err)
			}
		})
	}
}

func TestWriteNoteRejectsSymlinkEscape(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	vaultDir := filepath.Join(root, "vault")
	outsideDir := filepath.Join(root, "outside")

	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(vaultDir, "escape")
	if err := os.Symlink(outsideDir, linkPath); err != nil {
		t.Skipf("symlink not supported in this environment: %v", err)
	}

	v := New(vaultDir)
	_, _, err := v.WriteNoteHandler(ctx, nil, WriteNoteArgs{
		Path:    "escape/hijack.md",
		Content: "should not write outside vault",
	})
	if err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
	if !strings.Contains(err.Error(), "within vault") {
		t.Fatalf("expected vault-bound error, got: %v", err)
	}
}
