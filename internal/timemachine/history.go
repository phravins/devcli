package timemachine

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Commit represents a Git commit with metadata
type Commit struct {
	Hash         string
	ShortHash    string
	Author       string
	AuthorEmail  string
	Date         time.Time
	Message      string
	FilesChanged []string
	LinesAdded   int
	LinesRemoved int
	Diff         string
}

// GetFileHistory retrieves the complete commit history for a specific file
func GetFileHistory(repoPath, filePath string) ([]Commit, error) {
	// Use --follow to track file renames
	// Format: hash|short_hash|author_name|author_email|timestamp|commit_message
	cmd := exec.Command("git", "log", "--follow", "--pretty=format:%H|%h|%an|%ae|%at|%s", "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 6)
		if len(parts) != 6 {
			continue
		}

		timestamp, _ := strconv.ParseInt(parts[4], 10, 64)

		commit := Commit{
			Hash:        parts[0],
			ShortHash:   parts[1],
			Author:      parts[2],
			AuthorEmail: parts[3],
			Date:        time.Unix(timestamp, 0),
			Message:     parts[5],
		}

		commits = append(commits, commit)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing git log: %w", err)
	}

	// Get diff stats for each commit
	for i := range commits {
		diff, stats, err := getCommitDiff(repoPath, commits[i].Hash, filePath)
		if err == nil {
			commits[i].Diff = diff
			commits[i].LinesAdded = stats.Added
			commits[i].LinesRemoved = stats.Removed
		}
	}

	return commits, nil
}

// DiffStats contains statistics about changes in a commit
type DiffStats struct {
	Added   int
	Removed int
}

// getCommitDiff retrieves the diff for a specific commit and file
func getCommitDiff(repoPath, commitHash, filePath string) (string, DiffStats, error) {
	cmd := exec.Command("git", "show", "--format=", commitHash, "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", DiffStats{}, fmt.Errorf("git show failed: %w", err)
	}

	diff := string(output)
	stats := parseDiffStats(diff)

	return diff, stats, nil
}

// parseDiffStats calculates lines added/removed from a diff
func parseDiffStats(diff string) DiffStats {
	var stats DiffStats
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case '+':
			if !strings.HasPrefix(line, "+++") {
				stats.Added++
			}
		case '-':
			if !strings.HasPrefix(line, "---") {
				stats.Removed++
			}
		}
	}

	return stats
}

// GetCommitDetails retrieves detailed information about a specific commit
func GetCommitDetails(repoPath, commitHash string) (*Commit, error) {
	// Get commit metadata
	cmd := exec.Command("git", "show", "--format=%H|%h|%an|%ae|%at|%s|%b", "--no-patch", commitHash)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no commit data found")
	}

	parts := strings.SplitN(lines[0], "|", 7)
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid commit format")
	}

	timestamp, _ := strconv.ParseInt(parts[4], 10, 64)

	commit := &Commit{
		Hash:        parts[0],
		ShortHash:   parts[1],
		Author:      parts[2],
		AuthorEmail: parts[3],
		Date:        time.Unix(timestamp, 0),
		Message:     parts[5],
	}

	// Add full message body if exists
	if len(parts) > 6 && parts[6] != "" {
		commit.Message = parts[5] + "\n\n" + parts[6]
	}

	// Get files changed
	cmd = exec.Command("git", "show", "--name-only", "--format=", commitHash)
	cmd.Dir = repoPath

	output, err = cmd.Output()
	if err == nil {
		files := strings.Split(strings.TrimSpace(string(output)), "\n")
		commit.FilesChanged = files
	}

	return commit, nil
}

// GetDiffBetween retrieves the diff between two commits for a specific file
func GetDiffBetween(repoPath, hash1, hash2, filePath string) (string, error) {
	cmd := exec.Command("git", "diff", hash1, hash2, "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	return string(output), nil
}

// IsGitRepository checks if a directory is a Git repository
func IsGitRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path

	return cmd.Run() == nil
}

// GetRepositoryRoot finds the root directory of the Git repository
func GetRepositoryRoot(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
