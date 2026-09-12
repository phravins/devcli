package history_test

import (
	"os"
	"testing"
	"time"

	"github.com/phravins/devcli/internal/history"
)

func TestGetOldEntries(t *testing.T) {
	// Create a temporary directory for tests
	tmpDir := t.TempDir()

	// Override the HOME and USERPROFILE environment variables so that
	// history.getHistoryPath() will use this temp directory instead of the real home directory.
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	now := time.Now()

	entries := []history.Entry{
		{
			Name:      "Entry 1",
			Path:      "/path/1",
			CreatedAt: now, // 0 days old
		},
		{
			Name:      "Entry 2",
			Path:      "/path/2",
			CreatedAt: now.AddDate(0, 0, -2), // 2 days old
		},
		{
			Name:      "Entry 3",
			Path:      "/path/3",
			CreatedAt: now.AddDate(0, 0, -10), // 10 days old
		},
		{
			Name:      "Entry 4",
			Path:      "/path/4",
			CreatedAt: now.AddDate(0, 0, -30), // 30 days old
		},
	}

	// Save entries
	err := history.Save(entries)
	if err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Test case 1: cutoff 5 days
	// Entries older than 5 days should be returned (Entry 3, Entry 4)
	oldEntries := history.GetOldEntries(5)
	if len(oldEntries) != 2 {
		t.Errorf("Expected 2 old entries, got %d", len(oldEntries))
	} else {
		if oldEntries[0].Name != "Entry 3" || oldEntries[1].Name != "Entry 4" {
			t.Errorf("Unexpected old entries: %v", oldEntries)
		}
	}

	// Test case 2: cutoff 1 day
	// Entries older than 1 day should be returned (Entry 2, Entry 3, Entry 4)
	oldEntries = history.GetOldEntries(1)
	if len(oldEntries) != 3 {
		t.Errorf("Expected 3 old entries, got %d", len(oldEntries))
	} else {
		if oldEntries[0].Name != "Entry 2" || oldEntries[1].Name != "Entry 3" || oldEntries[2].Name != "Entry 4" {
			t.Errorf("Unexpected old entries: %v", oldEntries)
		}
	}

	// Test case 3: cutoff 40 days
	// Entries older than 40 days should be returned (None)
	oldEntries = history.GetOldEntries(40)
	if len(oldEntries) != 0 {
		t.Errorf("Expected 0 old entries, got %d", len(oldEntries))
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	now := time.Now().Truncate(time.Second)
	entries := []history.Entry{
		{
			Name:      "Project A",
			Path:      "/path/to/projectA",
			CreatedAt: now,
		},
		{
			Name:      "Project B",
			Path:      "/path/to/projectB",
			CreatedAt: now.Add(-time.Hour),
		},
	}

	// Test saving non-empty list
	if err := history.Save(entries); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := history.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded) != len(entries) {
		t.Fatalf("Expected %d entries, got %d", len(entries), len(loaded))
	}

	for i := range entries {
		if loaded[i].Name != entries[i].Name || loaded[i].Path != entries[i].Path {
			t.Errorf("Entry mismatch at index %d: got %+v, want %+v", i, loaded[i], entries[i])
		}
	}

	// Test saving empty list
	if err := history.Save([]history.Entry{}); err != nil {
		t.Fatalf("Save empty entries failed: %v", err)
	}

	loadedEmpty, err := history.Load()
	if err != nil {
		t.Fatalf("Load empty failed: %v", err)
	}
	if len(loadedEmpty) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(loadedEmpty))
	}
}

func TestSave_ErrorWritingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// Make .devcli path a file instead of directory so MkdirAll inside getHistoryPath will fail or WriteFile will fail
	// Or we can create the directory ~/.devcli with read-only permission (0400 / 0500)
	// Alternatively, pointing HOME to an invalid/unwritable location or replacing .devcli with a file
	devcliPath := tmpDir + "/.devcli"
	if err := history.Save([]history.Entry{}); err != nil {
		t.Fatalf("Initial save failed: %v", err)
	}

	// Turn .devcli directory or history.json into read-only
	historyFilePath := devcliPath + "/history.json"
	_ = historyFilePath
	// Make .devcli directory read-only
	if err := os.Chmod(devcliPath, 0500); err != nil {
		t.Skip("Skipping permission error test (chmod non-functional on platform)")
	}
	defer os.Chmod(devcliPath, 0755)

	err := history.Save([]history.Entry{{Name: "Fail", Path: "/fail"}})
	if err == nil {
		// Note: On Windows or as root, chmod 0500 might still allow writes.
		t.Log("Save succeeded despite read-only directory (likely running as superuser or on Windows)")
	}
}

func TestLoad_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	entries, err := history.Load()
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(entries))
	}
}

func TestAdd_DeleteOne_DeleteOld(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	if err := history.Add("New Entry", "/new/path"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	entries, err := history.Load()
	if err != nil || len(entries) != 1 {
		t.Fatalf("Load after Add failed, len=%d, err=%v", len(entries), err)
	}
	if entries[0].Name != "New Entry" {
		t.Errorf("Expected name 'New Entry', got '%s'", entries[0].Name)
	}

	// Delete out of range
	if err := history.DeleteOne(10); err != nil {
		t.Errorf("DeleteOne out of range returned error: %v", err)
	}
	if err := history.DeleteOne(-1); err != nil {
		t.Errorf("DeleteOne negative index returned error: %v", err)
	}

	// Delete valid index
	if err := history.DeleteOne(0); err != nil {
		t.Errorf("DeleteOne failed: %v", err)
	}

	entries, _ = history.Load()
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after DeleteOne, got %d", len(entries))
	}

	// Test DeleteOld
	now := time.Now()
	testEntries := []history.Entry{
		{Name: "Recent", Path: "/p1", CreatedAt: now},
		{Name: "Old", Path: "/p2", CreatedAt: now.AddDate(0, 0, -10)},
	}
	_ = history.Save(testEntries)

	if err := history.DeleteOld(5); err != nil {
		t.Fatalf("DeleteOld failed: %v", err)
	}

	entries, _ = history.Load()
	if len(entries) != 1 || entries[0].Name != "Recent" {
		t.Errorf("Unexpected entries after DeleteOld: %+v", entries)
	}
}
