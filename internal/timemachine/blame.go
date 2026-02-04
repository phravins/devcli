package timemachine

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// BlameLine represents a single line with its Git blame information
type BlameLine struct {
	LineNumber     int
	Content        string
	CommitHash     string
	Author         string
	AuthorEmail    string
	Timestamp      time.Time
	CommitMessage  string
	BoundaryCommit bool // First commit for this file
}

// GetBlame retrieves the Git blame information for a file
func GetBlame(repoPath, filePath string) ([]BlameLine, error) {
	// Run git blame with porcelain format
	cmd := exec.Command("git", "blame", "--line-porcelain", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame failed: %w", err)
	}

	return parseBlameOutput(string(output))
}

// parseBlameOutput parses the porcelain format output from git blame
func parseBlameOutput(output string) ([]BlameLine, error) {
	var lines []BlameLine
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentLine BlameLine
	var lineNum int

	for scanner.Scan() {
		text := scanner.Text()

		// Commit hash line (starts hash, original line num, final line num)
		if len(text) > 40 && text[40] == ' ' {
			// Save previous line if exists
			if currentLine.CommitHash != "" {
				lines = append(lines, currentLine)
			}

			// Parse new commit line
			parts := strings.Fields(text)
			currentLine = BlameLine{
				CommitHash: parts[0],
			}

			// Final line number is third field
			if len(parts) >= 3 {
				lineNum, _ = strconv.Atoi(parts[2])
				currentLine.LineNumber = lineNum
			}

			// Check if boundary commit (prefixed with ^)
			if strings.HasPrefix(parts[0], "^") {
				currentLine.BoundaryCommit = true
				currentLine.CommitHash = strings.TrimPrefix(parts[0], "^")
			}

		} else if strings.HasPrefix(text, "author ") {
			currentLine.Author = strings.TrimPrefix(text, "author ")

		} else if strings.HasPrefix(text, "author-mail ") {
			email := strings.TrimPrefix(text, "author-mail ")
			currentLine.AuthorEmail = strings.Trim(email, "<>")

		} else if strings.HasPrefix(text, "author-time ") {
			timeStr := strings.TrimPrefix(text, "author-time ")
			if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
				currentLine.Timestamp = time.Unix(timestamp, 0)
			}

		} else if strings.HasPrefix(text, "summary ") {
			currentLine.CommitMessage = strings.TrimPrefix(text, "summary ")

		} else if strings.HasPrefix(text, "\t") {
			// This is the actual line content
			currentLine.Content = strings.TrimPrefix(text, "\t")
		}
	}

	// Add the last line
	if currentLine.CommitHash != "" {
		lines = append(lines, currentLine)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing blame output: %w", err)
	}

	return lines, nil
}

// GetLineBlame retrieves blame info for a specific line in a file
func GetLineBlame(repoPath, filePath string, lineNum int) (*BlameLine, error) {
	cmd := exec.Command("git", "blame", "-L", fmt.Sprintf("%d,%d", lineNum, lineNum), "--line-porcelain", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame failed: %w", err)
	}

	lines, err := parseBlameOutput(string(output))
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no blame information found for line %d", lineNum)
	}

	return &lines[0], nil
}

// GetBlameRange retrieves blame info for a range of lines
func GetBlameRange(repoPath, filePath string, startLine, endLine int) ([]BlameLine, error) {
	cmd := exec.Command("git", "blame", "-L", fmt.Sprintf("%d,%d", startLine, endLine), "--line-porcelain", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame failed: %w", err)
	}

	return parseBlameOutput(string(output))
}
