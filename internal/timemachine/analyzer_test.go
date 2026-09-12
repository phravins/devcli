package timemachine

import (
	"reflect"
	"testing"
	"time"
)

func TestContainsAny(t *testing.T) {
	tests := []struct {
		s        string
		keywords []string
		expected bool
	}{
		{"fix issue in login", []string{"fix", "hotfix"}, true},
		{"add feature", []string{"fix", "hotfix"}, false},
		{"WIP: working on feature", []string{"wip", "todo", "temp"}, true},
		{"TODO: add error handling", []string{"wip", "todo", "temp"}, true},
		{"temp fix for crash", []string{"wip", "todo", "temp"}, true},
		{"clean code", []string{"wip", "todo", "temp"}, false},
	}

	for _, tt := range tests {
		result := containsAny(tt.s, tt.keywords...)
		if result != tt.expected {
			t.Errorf("containsAny(%q, %v) = %v; want %v", tt.s, tt.keywords, result, tt.expected)
		}
	}
}

func TestAnalyzeBugRisks(t *testing.T) {
	baseTime := time.Date(2023, 10, 25, 12, 0, 0, 0, time.UTC) // Wednesday 12 PM

	commits := []Commit{
		{
			Hash:         "hash1",
			ShortHash:    "h1",
			Author:       "Alice",
			Date:         baseTime,
			Message:      "Normal commit",
			LinesAdded:   10,
			LinesRemoved: 10,
		},
		{
			Hash:         "hash2",
			ShortHash:    "h2",
			Author:       "Bob",
			Date:         baseTime.Add(1 * time.Hour),
			Message:      "WIP: unfinished work",
			LinesAdded:   50,
			LinesRemoved: 20,
		},
		{
			Hash:         "hash3",
			ShortHash:    "h3",
			Author:       "Charlie",
			Date:         time.Date(2023, 10, 27, 18, 0, 0, 0, time.UTC), // Friday 6 PM
			Message:      "fix: patch critical bug",
			LinesAdded:   150,
			LinesRemoved: 100,
			FilesChanged: []string{"a", "b", "c", "d", "e", "f"},
		},
		{
			Hash:         "hash4",
			ShortHash:    "h4",
			Author:       "Dave",
			Date:         time.Date(2023, 10, 28, 2, 0, 0, 0, time.UTC), // Late night 2 AM
			Message:      "quick temp hack",
			LinesAdded:   5,
			LinesRemoved: 5,
		},
	}

	suspects := AnalyzeBugRisks(commits)

	if len(suspects) == 0 {
		t.Fatalf("expected bug suspects, got 0")
	}

	// Normal commit should not be in suspects (risk = 0.0 < 0.3)
	for _, s := range suspects {
		if s.Commit.Hash == "hash1" {
			t.Errorf("did not expect normal commit (hash1) to be flagged as suspect")
		}
	}

	// Verify hash2 (WIP message -> risk += 0.4)
	var foundHash2 bool
	for _, s := range suspects {
		if s.Commit.Hash == "hash2" {
			foundHash2 = true
			if s.Risk < 0.4 {
				t.Errorf("expected hash2 risk >= 0.4, got %f", s.Risk)
			}
		}
	}
	if !foundHash2 {
		t.Errorf("expected hash2 to be flagged as suspect")
	}

	// Verify hash3 (Friday evening + fix keyword + large refactor + multiple files -> capped at 1.0)
	var foundHash3 bool
	for _, s := range suspects {
		if s.Commit.Hash == "hash3" {
			foundHash3 = true
			if s.Risk != 1.0 {
				t.Errorf("expected hash3 risk capped at 1.0, got %f", s.Risk)
			}
		}
	}
	if !foundHash3 {
		t.Errorf("expected hash3 to be flagged as suspect")
	}
}

func TestGetRiskLevelAndColor(t *testing.T) {
	tests := []struct {
		risk          float64
		expectedLevel string
		expectedColor string
	}{
		{0.8, "High", "#FF4444"},
		{0.7, "High", "#FF4444"},
		{0.5, "Medium", "#FFA500"},
		{0.4, "Medium", "#FFA500"},
		{0.2, "Low", "#90EE90"},
		{0.0, "Low", "#90EE90"},
	}

	for _, tt := range tests {
		level := GetRiskLevel(tt.risk)
		if level != tt.expectedLevel {
			t.Errorf("GetRiskLevel(%f) = %q; want %q", tt.risk, level, tt.expectedLevel)
		}
		color := GetRiskColor(tt.risk)
		if color != tt.expectedColor {
			t.Errorf("GetRiskColor(%f) = %q; want %q", tt.risk, color, tt.expectedColor)
		}
	}
}

func TestFindSuspiciousLines(t *testing.T) {
	blameLines := []BlameLine{
		{LineNumber: 1, CommitHash: "hash1"},
		{LineNumber: 2, CommitHash: "hash2"},
		{LineNumber: 3, CommitHash: "hash3"},
		{LineNumber: 4, CommitHash: "hash1"},
	}

	suspects := []BugSuspect{
		{Commit: Commit{Hash: "hash2"}},
		{Commit: Commit{Hash: "hash3"}},
	}

	suspicious := FindSuspiciousLines(blameLines, suspects)
	expected := []int{2, 3}

	if !reflect.DeepEqual(suspicious, expected) {
		t.Errorf("FindSuspiciousLines() = %v; want %v", suspicious, expected)
	}
}

func TestAnalyzeChurn(t *testing.T) {
	commits := []Commit{
		{Hash: "hash1"},
	}
	churn := AnalyzeChurn(commits)
	if churn == nil {
		t.Errorf("expected non-nil map from AnalyzeChurn")
	}
}
