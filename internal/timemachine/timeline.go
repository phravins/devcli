package timemachine

import (
	"fmt"
	"strings"
)

type Timeline struct {
	RepoPath	string
	FilePath	string
	Commits		[]Commit
	CurrentIndex	int
	BlameData	[]BlameLine
}

func NewTimeline(repoPath, filePath string) (*Timeline, error) {

	if !IsGitRepository(repoPath) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	commits, err := GetFileHistory(repoPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commit history found for file: %s", filePath)
	}

	blame, err := GetBlame(repoPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get blame data: %w", err)
	}

	return &Timeline{
		RepoPath:	repoPath,
		FilePath:	filePath,
		Commits:	commits,
		CurrentIndex:	0,
		BlameData:	blame,
	}, nil
}

func (t *Timeline) Next() error {
	if t.CurrentIndex >= len(t.Commits)-1 {
		return fmt.Errorf("already at oldest commit")
	}

	t.CurrentIndex++
	return t.updateBlameData()
}

func (t *Timeline) Previous() error {
	if t.CurrentIndex <= 0 {
		return fmt.Errorf("already at newest commit")
	}

	t.CurrentIndex--
	return t.updateBlameData()
}

func (t *Timeline) MoveTo(commitHash string) error {
	for i, commit := range t.Commits {
		if commit.Hash == commitHash || commit.ShortHash == commitHash {
			t.CurrentIndex = i
			return t.updateBlameData()
		}
	}

	return fmt.Errorf("commit not found: %s", commitHash)
}

func (t *Timeline) MoveToIndex(index int) error {
	if index < 0 || index >= len(t.Commits) {
		return fmt.Errorf("index out of range: %d", index)
	}

	t.CurrentIndex = index
	return t.updateBlameData()
}

func (t *Timeline) updateBlameData() error {
	currentHash := t.Commits[t.CurrentIndex].Hash

	blame, err := GetBlameAtCommit(t.RepoPath, t.FilePath, currentHash)
	if err != nil {
		return fmt.Errorf("failed to get blame at commit %s: %w", currentHash, err)
	}

	t.BlameData = blame
	return nil
}

func (t *Timeline) GetCurrentCommit() *Commit {
	if t.CurrentIndex < 0 || t.CurrentIndex >= len(t.Commits) {
		return nil
	}
	return &t.Commits[t.CurrentIndex]
}

func (t *Timeline) GetDiffToCurrent(fromHash string) (string, error) {
	currentHash := t.Commits[t.CurrentIndex].Hash
	return GetDiffBetween(t.RepoPath, fromHash, currentHash, t.FilePath)
}

func (t *Timeline) GetProgress() float64 {
	if len(t.Commits) <= 1 {
		return 1.0
	}
	return float64(t.CurrentIndex) / float64(len(t.Commits)-1)
}

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

func (t *Timeline) GetCommitsByAuthor(author string) []Commit {
	var filtered []Commit

	for _, commit := range t.Commits {
		if strings.EqualFold(commit.Author, author) {
			filtered = append(filtered, commit)
		}
	}

	return filtered
}

func (t *Timeline) GetCommitCount() int {
	return len(t.Commits)
}

func (t *Timeline) HasNext() bool {
	return t.CurrentIndex < len(t.Commits)-1
}

func (t *Timeline) HasPrevious() bool {
	return t.CurrentIndex > 0
}
