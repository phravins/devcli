package fileops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchFiles_PathSanitization(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	testFile := filepath.Join(subDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("target pattern here"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test search with redundant dot/slash path
	dirtyPath := subDir + "/./"
	matches, err := searchFiles("pattern", dirtyPath)
	if err != nil {
		t.Fatalf("searchFiles failed: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestBulkRename_PathSanitization(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	testFile := filepath.Join(subDir, "old_name.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test bulkRename with redundant relative path components
	dirtyPath := tmpDir + "/sub/../sub/."
	count, err := bulkRename("old", "new", dirtyPath)
	if err != nil {
		t.Fatalf("bulkRename failed: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 renamed file, got %d", count)
	}

	renamedFile := filepath.Join(subDir, "new_name.txt")
	if _, err := os.Stat(renamedFile); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", renamedFile)
	}
}

func TestBulkRename_PathTraversalPrevention(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "target_file.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Attempt path traversal replacement
	_, err := bulkRename("target", "../escaped", tmpDir)
	if err == nil {
		t.Fatalf("expected path traversal error, got nil")
	}
}
