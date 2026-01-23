package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

type LocalHFProvider struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    *bufio.Scanner
	mutex     sync.Mutex
	isRunning bool
}

func (p *LocalHFProvider) Name() string {
	return "Local (Python/HF)"
}

func (p *LocalHFProvider) Model() string {
	return "local-model"
}

func (p *LocalHFProvider) Configure(cfg *config.Config) error {
	_, err := exec.LookPath("python")
	if err != nil {
		return fmt.Errorf("python is not installed. Required for local HF models")
	}
	return nil
}

func (p *LocalHFProvider) IsLocal() bool {
	return true
}

func (p *LocalHFProvider) ensureStarted() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isRunning {
		return nil
	}
	cwd, _ := os.Getwd()
	scriptPath := filepath.Join(cwd, "internal", "ai", "scripts", "hf_chat.py")

	cmd := exec.Command("python", "-u", scriptPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start python script: %w", err)
	}

	p.cmd = cmd
	p.stdin = stdin
	p.stdout = bufio.NewScanner(stdout)
	p.isRunning = true

	return nil
}

type pythonRequest struct {
	Prompt string `json:"prompt"`
}

type pythonResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func (p *LocalHFProvider) Send(messages []ai.Message) (string, error) {
	if err := p.ensureStarted(); err != nil {
		return "", err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	var prompt string
	for _, m := range messages {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}
	prompt += "assistant:"

	req := pythonRequest{Prompt: prompt}
	jsonBytes, _ := json.Marshal(req)
	jsonBytes = append(jsonBytes, '\n')

	if _, err := p.stdin.Write(jsonBytes); err != nil {
		p.isRunning = false
		return "", fmt.Errorf("failed to write to python process: %w", err)
	}
	if p.stdout.Scan() {
		line := p.stdout.Bytes()
		var resp pythonResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			return "", fmt.Errorf("invalid json from python: %s", line)
		}
		if resp.Error != "" {
			return "", fmt.Errorf("python error: %s", resp.Error)
		}
		return resp.Response, nil
	}

	if err := p.stdout.Err(); err != nil {
		p.isRunning = false
		return "", err
	}

	return "", fmt.Errorf("python process closed stream")
}
