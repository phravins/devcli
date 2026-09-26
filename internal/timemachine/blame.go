package timemachine

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type BlameLine struct {
	LineNumber	int
	Content		string
	CommitHash	string
	Author		string
	AuthorEmail	string
	Timestamp	time.Time
	CommitMessage	string
	BoundaryCommit	bool
}

func GetBlame(repoPath, filePath string) ([]BlameLine, error) {

	cmd := exec.Command("git", "blame", "--line-porcelain", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame failed: %w", err)
	}

	return parseBlameOutput(string(output))
}

func GetBlameAtCommit(repoPath, filePath, commitHash string) ([]BlameLine, error) {

	cmd := exec.Command("git", "blame", "--line-porcelain", commitHash, "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame failed at commit %s: %w", commitHash, err)
	}

	return parseBlameOutput(string(output))
}

func parseBlameOutput(output string) ([]BlameLine, error) {
	var lines []BlameLine
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentLine BlameLine
	var lineNum int

	for scanner.Scan() {
		text := scanner.Text()

		if len(text) > 40 && text[40] == ' ' {

			if currentLine.CommitHash != "" {
				lines = append(lines, currentLine)
			}

			parts := strings.Fields(text)
			currentLine = BlameLine{
				CommitHash: parts[0],
			}

			if len(parts) >= 3 {
				lineNum, _ = strconv.Atoi(parts[2])
				currentLine.LineNumber = lineNum
			}

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

			currentLine.Content = strings.TrimPrefix(text, "\t")
		}
	}

	if currentLine.CommitHash != "" {
		lines = append(lines, currentLine)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing blame output: %w", err)
	}

	return lines, nil
}

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

func GetBlameRange(repoPath, filePath string, startLine, endLine int) ([]BlameLine, error) {
	cmd := exec.Command("git", "blame", "-L", fmt.Sprintf("%d,%d", startLine, endLine), "--line-porcelain", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame failed: %w", err)
	}

	return parseBlameOutput(string(output))
}
