package tui

import (
	"fmt"
	"testing"
)

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s           string
		lowerSubstr string
		expected    bool
	}{
		{"/usr/local/main.go", "main", true},
		{"/USR/LOCAL/MAIN.GO", "main", true},
		{"/usr/local/main.go", "MAIN", false}, // lowerSubstr must be lowercased by caller
		{"/usr/local/main.go", "xyz", false},
		{"/usr/local/main.go", "", true},
		{"short", "longerstring", false},
		{"Café/main.go", "café", true},
		{"CAFÉ/MAIN.GO", "café", true},
		{"file_123_component.go", "component", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.s, tt.lowerSubstr), func(t *testing.T) {
			got := containsIgnoreCase(tt.s, tt.lowerSubstr)
			if got != tt.expected {
				t.Errorf("containsIgnoreCase(%q, %q) = %v; want %v", tt.s, tt.lowerSubstr, got, tt.expected)
			}
		})
	}
}

func TestMatchPaths(t *testing.T) {
	t.Run("empty query", func(t *testing.T) {
		res := matchPaths([]string{"a.go", "b.go"}, "")
		if res != nil {
			t.Errorf("expected nil for empty query, got %v", res)
		}
	})

	t.Run("fuzzy path matching", func(t *testing.T) {
		paths := []string{"/app/main.go", "/app/utils.go", "/app/config.yaml"}
		res := matchPaths(paths, "main")
		if len(res) != 1 {
			t.Fatalf("expected 1 result, got %d", len(res))
		}
		if res[0].Name() != "/app/main.go" {
			t.Errorf("expected /app/main.go, got %s", res[0].Name())
		}
	})

	t.Run("fast path matching and result capping", func(t *testing.T) {
		count := 6000
		paths := make([]string, count)
		for i := 0; i < count; i++ {
			paths[i] = fmt.Sprintf("/project/path_%d_item.go", i)
		}

		res := matchPaths(paths, "item")
		if len(res) != maxSearchResults {
			t.Errorf("expected capped results of %d, got %d", maxSearchResults, len(res))
		}
	})
}

func TestFilterFiles(t *testing.T) {
	m := FileManagerModel{
		allFilePaths: []string{"/a/first.go", "/a/second.py", "/b/third.js"},
		globalSearch: true,
	}

	m.filterFiles("second")
	if len(m.filtered) != 1 {
		t.Fatalf("expected 1 filtered file, got %d", len(m.filtered))
	}
	if m.filtered[0].Name() != "/a/second.py" {
		t.Errorf("expected /a/second.py, got %s", m.filtered[0].Name())
	}

	m.filterFiles("")
	if len(m.filtered) != len(m.files) {
		t.Errorf("expected resetting filtered to m.files on empty query")
	}
}
