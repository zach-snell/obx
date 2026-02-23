package vault

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDeleteNoteDryRunDoesNotDelete(t *testing.T) {
	ctx := context.Background()
	v, dir := setupTestVault(t)
	writeTestFile(t, dir, "keep.md", "hello")

	_, _, err := v.DeleteNoteHandler(ctx, nil, DeleteNoteArgs{
		Path:   "keep.md",
		DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "keep.md")); err != nil {
		t.Fatalf("expected file to remain after dry_run, stat err: %v", err)
	}
}

func TestBatchEditDryRunDoesNotWrite(t *testing.T) {
	ctx := context.Background()
	v, dir := setupTestVault(t)
	writeTestFile(t, dir, "doc.md", "hello world")

	_, _, err := v.BatchEditNoteHandler(ctx, nil, BatchEditArgs{
		Path:   "doc.md",
		DryRun: true,
		Edits: []EditEntry{
			{OldText: "hello", NewText: "hi"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	got := readTestFile(t, dir, "doc.md")
	if got != "hello world" {
		t.Fatalf("expected original content unchanged, got: %q", got)
	}
}

func TestBulkMoveDryRunDoesNotMove(t *testing.T) {
	ctx := context.Background()
	v, dir := setupTestVault(t)
	writeTestFile(t, dir, "a.md", "A")

	_, _, err := v.BulkMoveHandler(ctx, nil, BulkMoveArgs{
		Paths:       "a.md",
		Destination: "archive",
		DryRun:      true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "a.md")); err != nil {
		t.Fatalf("expected source to remain after dry_run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "archive", "a.md")); !os.IsNotExist(err) {
		t.Fatalf("expected destination not to be created in dry_run")
	}
}

func TestMergeNotesDryRunDoesNotWriteOrDelete(t *testing.T) {
	ctx := context.Background()
	v, dir := setupTestVault(t)
	writeTestFile(t, dir, "one.md", "One")
	writeTestFile(t, dir, "two.md", "Two")

	_, _, err := v.MergeNotesHandler(ctx, nil, MergeNotesArgs{
		Paths:           "one.md,two.md",
		Output:          "merged.md",
		DeleteOriginals: true,
		DryRun:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "merged.md")); !os.IsNotExist(err) {
		t.Fatalf("expected merged output not to be created in dry_run")
	}
	if _, err := os.Stat(filepath.Join(dir, "one.md")); err != nil {
		t.Fatalf("expected source one.md to remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "two.md")); err != nil {
		t.Fatalf("expected source two.md to remain: %v", err)
	}
}

func TestWriteNoteExpectedMtimeMismatch(t *testing.T) {
	ctx := context.Background()
	v, dir := setupTestVault(t)
	writeTestFile(t, dir, "doc.md", "v1")

	stale := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano)
	_, _, err := v.WriteNoteHandler(ctx, nil, WriteNoteArgs{
		Path:          "doc.md",
		Content:       "v2",
		ExpectedMtime: stale,
	})
	if err == nil {
		t.Fatal("expected mtime mismatch error")
	}
	if !strings.Contains(err.Error(), "mtime mismatch") {
		t.Fatalf("expected mtime mismatch error, got: %v", err)
	}

	got := readTestFile(t, dir, "doc.md")
	if got != "v1" {
		t.Fatalf("expected file unchanged after mismatch, got: %q", got)
	}
}

func TestEditNoteExpectedMtimeMatch(t *testing.T) {
	ctx := context.Background()
	v, dir := setupTestVault(t)
	writeTestFile(t, dir, "doc.md", "hello world")

	mtime, err := fileMtimeRFC3339Nano(filepath.Join(dir, "doc.md"))
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = v.EditNoteHandler(ctx, nil, EditNoteArgs{
		Path:          "doc.md",
		OldText:       "hello",
		NewText:       "hi",
		ExpectedMtime: mtime,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := readTestFile(t, dir, "doc.md")
	if got != "hi world" {
		t.Fatalf("expected edited content, got: %q", got)
	}
}
