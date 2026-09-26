package history_test

import (
	"testing"
	"time"

	"github.com/phravins/devcli/internal/history"
)

func TestGetOldEntries(t *testing.T) {

	tmpDir := t.TempDir()

	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	now := time.Now()

	entries := []history.Entry{
		{
			Name:		"Entry 1",
			Path:		"/path/1",
			CreatedAt:	now,
		},
		{
			Name:		"Entry 2",
			Path:		"/path/2",
			CreatedAt:	now.AddDate(0, 0, -2),
		},
		{
			Name:		"Entry 3",
			Path:		"/path/3",
			CreatedAt:	now.AddDate(0, 0, -10),
		},
		{
			Name:		"Entry 4",
			Path:		"/path/4",
			CreatedAt:	now.AddDate(0, 0, -30),
		},
	}

	err := history.Save(entries)
	if err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	oldEntries := history.GetOldEntries(5)
	if len(oldEntries) != 2 {
		t.Errorf("Expected 2 old entries, got %d", len(oldEntries))
	} else {
		if oldEntries[0].Name != "Entry 3" || oldEntries[1].Name != "Entry 4" {
			t.Errorf("Unexpected old entries: %v", oldEntries)
		}
	}

	oldEntries = history.GetOldEntries(1)
	if len(oldEntries) != 3 {
		t.Errorf("Expected 3 old entries, got %d", len(oldEntries))
	} else {
		if oldEntries[0].Name != "Entry 2" || oldEntries[1].Name != "Entry 3" || oldEntries[2].Name != "Entry 4" {
			t.Errorf("Unexpected old entries: %v", oldEntries)
		}
	}

	oldEntries = history.GetOldEntries(40)
	if len(oldEntries) != 0 {
		t.Errorf("Expected 0 old entries, got %d", len(oldEntries))
	}
}
