package timemachine

import (
	"strings"
	"time"
)

type BugSuspect struct {
	Commit	Commit
	Reason	string
	Risk	float64
	Context	[]string
}

func containsAny(s string, keywords ...string) bool {
	sLower := strings.ToLower(s)
	for _, kw := range keywords {
		if strings.Contains(sLower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func AnalyzeBugRisks(commits []Commit) []BugSuspect {
	var suspects []BugSuspect

	for i, commit := range commits {
		risk := 0.0
		reasons := []string{}
		context := []string{}

		hour := commit.Date.Hour()
		if hour >= 23 || hour <= 5 {
			risk += 0.3
			reasons = append(reasons, "Late-night commit")
		}

		totalChanges := commit.LinesAdded + commit.LinesRemoved
		if totalChanges > 200 {
			risk += 0.4
			reasons = append(reasons, "Large refactor")
		} else if totalChanges > 100 {
			risk += 0.2
			reasons = append(reasons, "Significant changes")
		}

		if containsAny(commit.Message, "fix", "hotfix", "patch", "bugfix") {
			risk += 0.3
			reasons = append(reasons, "Quick fix commit")
		}

		if containsAny(commit.Message, "wip", "todo", "temp") {
			risk += 0.4
			reasons = append(reasons, "Work in progress")
		}

		if i > 0 {
			nextCommit := commits[i-1]
			timeDiff := nextCommit.Date.Sub(commit.Date)

			if timeDiff < 2*time.Hour {
				if containsAny(nextCommit.Message, "fix") {
					risk += 0.3
					reasons = append(reasons, "Followed by quick fix")
					context = append(context, "Next commit: "+nextCommit.Message)
				}
			}
		}

		if len(commit.FilesChanged) > 5 {
			risk += 0.2
			reasons = append(reasons, "Multiple files changed")
		}

		if commit.Date.Weekday() == time.Friday && hour >= 16 {
			risk += 0.2
			reasons = append(reasons, "Friday evening commit")
		}

		if risk > 1.0 {
			risk = 1.0
		}

		if risk >= 0.3 {
			suspects = append(suspects, BugSuspect{
				Commit:		commit,
				Reason:		strings.Join(reasons, ", "),
				Risk:		risk,
				Context:	context,
			})
		}
	}

	return suspects
}

func GetRiskLevel(risk float64) string {
	if risk >= 0.7 {
		return "High"
	} else if risk >= 0.4 {
		return "Medium"
	}
	return "Low"
}

func GetRiskColor(risk float64) string {
	if risk >= 0.7 {
		return "#FF4444"
	} else if risk >= 0.4 {
		return "#FFA500"
	}
	return "#90EE90"
}

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

type ChurnData struct {
	LineNumber	int
	ChangeCount	int
	LastModified	time.Time
	Authors		[]string
}

func AnalyzeChurn(commits []Commit) map[int]ChurnData {

	churn := make(map[int]ChurnData)

	return churn
}
