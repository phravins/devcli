package history_test

import (
	"os"
	"path/filepath"
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

func TestLoad(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		entries, err := history.Load()
		if err != nil {
			t.Fatalf("Expected no error when file does not exist, got: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("Expected empty entries slice, got %d entries", len(entries))
		}
	})

	t.Run("valid history file", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		expected := []history.Entry{
			{Name: "Project A", Path: "/path/a"},
			{Name: "Project B", Path: "/path/b"},
		}
		if err := history.Save(expected); err != nil {
			t.Fatalf("Failed to save initial history: %v", err)
		}

		entries, err := history.Load()
		if err != nil {
			t.Fatalf("Expected no error loading valid file, got: %v", err)
		}
		if len(entries) != len(expected) {
			t.Fatalf("Expected %d entries, got %d", len(expected), len(entries))
		}
		if entries[0].Name != expected[0].Name || entries[1].Name != expected[1].Name {
			t.Errorf("Loaded entries do not match expected: %v", entries)
		}
	})

	t.Run("invalid json content", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		// Create .devcli dir and corrupt history.json
		devcliDir := filepath.Join(tmpDir, ".devcli")
		if err := os.MkdirAll(devcliDir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(devcliDir, "history.json"), []byte("{corrupt json"), 0644); err != nil {
			t.Fatalf("Failed to write corrupt json: %v", err)
		}

		entries, err := history.Load()
		if err != nil {
			t.Fatalf("Expected no error when unmarshaling fails, got: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("Expected empty entries slice on corrupt JSON, got %d entries", len(entries))
		}
	})

	t.Run("read file error", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		// Create history.json as a directory so os.ReadFile fails with read error (not IsNotExist)
		devcliDir := filepath.Join(tmpDir, ".devcli")
		if err := os.MkdirAll(filepath.Join(devcliDir, "history.json"), 0755); err != nil {
			t.Fatalf("Failed to create directory as history.json: %v", err)
		}

		entries, err := history.Load()
		if err == nil {
			t.Fatalf("Expected error when history.json is a directory, got nil")
		}
		if entries != nil {
			t.Errorf("Expected nil entries slice on read error, got %v", entries)
		}
	})
}
