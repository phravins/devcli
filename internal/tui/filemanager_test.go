package tui

import (
	"testing"
)

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"/path/to/File.go", "file", true},
		{"/path/to/File.go", "FILE", false}, // substr must be passed as lowerQuery
		{"/path/to/File.go", "path", true},
		{"/path/to/File.go", "xyz", false},
		{"/path/to/File.go", "", true},
		{"", "file", false},
		{"/path/to/ÜberFile.go", "über", true},
	}

	for _, tt := range tests {
		got := containsIgnoreCase(tt.s, tt.substr)
		if got != tt.expected {
			t.Errorf("containsIgnoreCase(%q, %q) = %v; want %v", tt.s, tt.substr, got, tt.expected)
		}
	}
}

func TestFilterFiles_MaxResults(t *testing.T) {
	paths := make([]string, 2000)
	for i := 0; i < 2000; i++ {
		paths[i] = "file_test.go"
	}

	m := FileManagerModel{
		allFilePaths: paths,
		globalSearch: true,
	}

	m.filterFiles("file")
	if len(m.filtered) != 1000 {
		t.Errorf("expected 1000 results max, got %d", len(m.filtered))
	}
}
