package timemachine

import (
	"fmt"
	"strings"
)

// Timeline manages the commit history and navigation for a file
type Timeline struct {
	RepoPath     string
	FilePath     string
	Commits      []Commit
	CurrentIndex int
	BlameData    []BlameLine
}

// NewTimeline creates a new timeline for a given file
func NewTimeline(repoPath, filePath string) (*Timeline, error) {
	// Verify it's a git repository
	if !IsGitRepository(repoPath) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Get file history
	commits, err := GetFileHistory(repoPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commit history found for file: %s", filePath)
	}

	// Get current blame data
	blame, err := GetBlame(repoPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get blame data: %w", err)
	}

	return &Timeline{
		RepoPath:     repoPath,
		FilePath:     filePath,
		Commits:      commits,
		CurrentIndex: 0, // Start at most recent
		BlameData:    blame,
	}, nil
}

// Next moves to the next commit in history (older)
func (t *Timeline) Next() error {
	if t.CurrentIndex >= len(t.Commits)-1 {
		return fmt.Errorf("already at oldest commit")
	}

	t.CurrentIndex++
	return t.updateBlameData()
}

// Previous moves to the previous commit in history (newer)
func (t *Timeline) Previous() error {
	if t.CurrentIndex <= 0 {
		return fmt.Errorf("already at newest commit")
	}

	t.CurrentIndex--
	return t.updateBlameData()
}

// MoveTo jumps to a specific commit by hash
func (t *Timeline) MoveTo(commitHash string) error {
	for i, commit := range t.Commits {
		if commit.Hash == commitHash || commit.ShortHash == commitHash {
			t.CurrentIndex = i
			return t.updateBlameData()
		}
	}

	return fmt.Errorf("commit not found: %s", commitHash)
}

// MoveToIndex jumps to a specific index in the timeline
func (t *Timeline) MoveToIndex(index int) error {
	if index < 0 || index >= len(t.Commits) {
		return fmt.Errorf("index out of range: %d", index)
	}

	t.CurrentIndex = index
	return t.updateBlameData()
}

// updateBlameData refreshes blame data for current commit
func (t *Timeline) updateBlameData() error {
	currentHash := t.Commits[t.CurrentIndex].Hash

	// Get blame at this specific commit
	// Note: Git blame at a specific commit shows state AT that commit
	blame, err := getBlameAtCommit(t.RepoPath, t.FilePath, currentHash)
	if err != nil {
		return fmt.Errorf("failed to get blame at commit %s: %w", currentHash, err)
	}

	t.BlameData = blame
	return nil
}

// getBlameAtCommit gets blame data for a file at a specific commit
func getBlameAtCommit(repoPath, filePath, commitHash string) ([]BlameLine, error) {
	// Use git blame <commit> -- <file> to get blame at that point in time
	// This is more complex - for MVP, we'll use current blame
	// TODO: Implement historical blame
	return GetBlame(repoPath, filePath)
}

// GetCurrentCommit returns the commit at the current timeline position
func (t *Timeline) GetCurrentCommit() *Commit {
	if t.CurrentIndex < 0 || t.CurrentIndex >= len(t.Commits) {
		return nil
	}
	return &t.Commits[t.CurrentIndex]
}

// GetDiffToCurrent gets the diff from a specific commit to current position
func (t *Timeline) GetDiffToCurrent(fromHash string) (string, error) {
	currentHash := t.Commits[t.CurrentIndex].Hash
	return GetDiffBetween(t.RepoPath, fromHash, currentHash, t.FilePath)
}

// GetProgress returns the current position as a percentage (0.0 to 1.0)
func (t *Timeline) GetProgress() float64 {
	if len(t.Commits) <= 1 {
		return 1.0
	}
	return float64(t.CurrentIndex) / float64(len(t.Commits)-1)
}

// GetAuthors returns a list of unique authors who contributed to this file
func (t *Timeline) GetAuthors() []string {
	authorMap := make(map[string]bool)

	for _, commit := range t.Commits {
		authorMap[commit.Author] = true
	}

	for _, line := range t.BlameData {
		authorMap[line.Author] = true
	}

	authors := make([]string, 0, len(authorMap))
	for author := range authorMap {
		authors = append(authors, author)
	}

	return authors
}

// GetCommitsByAuthor returns commits filtered by author
func (t *Timeline) GetCommitsByAuthor(author string) []Commit {
	var filtered []Commit

	for _, commit := range t.Commits {
		if strings.EqualFold(commit.Author, author) {
			filtered = append(filtered, commit)
		}
	}

	return filtered
}

// GetCommitCount returns the total number of commits
func (t *Timeline) GetCommitCount() int {
	return len(t.Commits)
}

// HasNext returns true if there's a next (older) commit
func (t *Timeline) HasNext() bool {
	return t.CurrentIndex < len(t.Commits)-1
}

// HasPrevious returns true if there's a previous (newer) commit
func (t *Timeline) HasPrevious() bool {
	return t.CurrentIndex > 0
}
