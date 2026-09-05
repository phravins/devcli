package history_test

import (
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
