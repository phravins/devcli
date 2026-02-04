package timemachine

import (
	"strings"
	"time"
)

// BugSuspect represents a commit that might have introduced a bug
type BugSuspect struct {
	Commit  Commit
	Reason  string
	Risk    float64  // 0.0 - 1.0, higher = more suspicious
	Context []string // Additional context
}

// AnalyzeBugRisks analyzes commits and identifies suspicious patterns
func AnalyzeBugRisks(commits []Commit) []BugSuspect {
	var suspects []BugSuspect

	for i, commit := range commits {
		risk := 0.0
		reasons := []string{}
		context := []string{}

		// Check for late-night commits (11 PM - 5 AM)
		hour := commit.Date.Hour()
		if hour >= 23 || hour <= 5 {
			risk += 0.3
			reasons = append(reasons, "Late-night commit")
		}

		// Check for large changes
		totalChanges := commit.LinesAdded + commit.LinesRemoved
		if totalChanges > 200 {
			risk += 0.4
			reasons = append(reasons, "Large refactor")
		} else if totalChanges > 100 {
			risk += 0.2
			reasons = append(reasons, "Significant changes")
		}

		// Check commit message for fix keywords
		msgLower := strings.ToLower(commit.Message)
		if strings.Contains(msgLower, "fix") ||
			strings.Contains(msgLower, "hotfix") ||
			strings.Contains(msgLower, "patch") ||
			strings.Contains(msgLower, "bugfix") {
			risk += 0.3
			reasons = append(reasons, "Quick fix commit")
		}

		// Check for WIP or TODO in message
		if strings.Contains(msgLower, "wip") ||
			strings.Contains(msgLower, "todo") ||
			strings.Contains(msgLower, "temp") {
			risk += 0.4
			reasons = append(reasons, "Work in progress")
		}

		// Check if commit was followed by a fix soon after
		if i > 0 {
			nextCommit := commits[i-1]
			timeDiff := nextCommit.Date.Sub(commit.Date)

			if timeDiff < 2*time.Hour {
				nextMsgLower := strings.ToLower(nextCommit.Message)
				if strings.Contains(nextMsgLower, "fix") {
					risk += 0.3
					reasons = append(reasons, "Followed by quick fix")
					context = append(context, "Next commit: "+nextCommit.Message)
				}
			}
		}

		// Check for multiple files changed (might introduce integration bugs)
		if len(commit.FilesChanged) > 5 {
			risk += 0.2
			reasons = append(reasons, "Multiple files changed")
		}

		// Check for Friday evening commits (technical debt territory)
		if commit.Date.Weekday() == time.Friday && hour >= 16 {
			risk += 0.2
			reasons = append(reasons, "Friday evening commit")
		}

		// Cap risk at 1.0
		if risk > 1.0 {
			risk = 1.0
		}

		// Only include if risk is above threshold
		if risk >= 0.3 {
			suspects = append(suspects, BugSuspect{
				Commit:  commit,
				Reason:  strings.Join(reasons, ", "),
				Risk:    risk,
				Context: context,
			})
		}
	}

	return suspects
}

// GetRiskLevel returns a human-readable risk level
func GetRiskLevel(risk float64) string {
	if risk >= 0.7 {
		return "High"
	} else if risk >= 0.4 {
		return "Medium"
	}
	return "Low"
}

// GetRiskColor returns a color code for the risk level
func GetRiskColor(risk float64) string {
	if risk >= 0.7 {
		return "#FF4444" // Red
	} else if risk >= 0.4 {
		return "#FFA500" // Orange
	}
	return "#90EE90" // Light green
}

// FindSuspiciousLines identifies lines that might be buggy based on blame data
func FindSuspiciousLines(blameLines []BlameLine, suspects []BugSuspect) []int {
	suspectHashes := make(map[string]bool)
	for _, suspect := range suspects {
		suspectHashes[suspect.Commit.Hash] = true
	}

	var suspiciousLines []int
	for _, line := range blameLines {
		if suspectHashes[line.CommitHash] {
			suspiciousLines = append(suspiciousLines, line.LineNumber)
		}
	}

	return suspiciousLines
}

// CalculateCodeChurn calculates how frequently each line has been modified
type ChurnData struct {
	LineNumber   int
	ChangeCount  int
	LastModified time.Time
	Authors      []string
}

// AnalyzeChurn analyzes code churn for a file
func AnalyzeChurn(commits []Commit) map[int]ChurnData {
	// Simplified churn analysis
	// For MVP, we'll count commits as a proxy for churn
	churn := make(map[int]ChurnData)

	// This would need more sophisticated diff parsing
	// For now, return empty - can be enhanced later

	return churn
}
