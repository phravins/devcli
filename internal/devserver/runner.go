package devserver

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"sync"
)

// ANSI escape code regex
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

type LogLine struct {
	ServerName string
	Line       string
	IsError    bool
}

type Runner struct {
	ctx       context.Context
	cancel    context.CancelFunc
	processes []*exec.Cmd
	logChan   chan LogLine
	wg        sync.WaitGroup
}

func NewRunner() *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		ctx:       ctx,
		cancel:    cancel,
		processes: make([]*exec.Cmd, 0),
		logChan:   make(chan LogLine, 100),
	}
}

func (r *Runner) Start(info ProjectInfo) error {
	if info.Type == TypeUnknown || len(info.Servers) == 0 {
		return fmt.Errorf("unable to detect project type or no servers configured")
	}

	for _, server := range info.Servers {
		if err := r.startServer(server); err != nil {
			r.Stop()
			return fmt.Errorf("failed to start %s: %w", server.Name, err)
		}
	}

	return nil
}

func (r *Runner) startServer(config ServerConfig) error {
	cmd := exec.CommandContext(r.ctx, config.Cmd, config.Args...)
	if config.Dir != "" {
		cmd.Dir = config.Dir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	r.processes = append(r.processes, cmd)

	// Stream stdout
	r.wg.Add(1)
	go r.streamLogs(config.Name, stdout, false)

	// Stream stderr
	r.wg.Add(1)
	go r.streamLogs(config.Name, stderr, true)

	return nil
}

func (r *Runner) streamLogs(serverName string, reader io.Reader, isError bool) {
	defer r.wg.Done()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		select {
		case <-r.ctx.Done():
			return
		case r.logChan <- LogLine{
			ServerName: serverName,
			Line:       re.ReplaceAllString(scanner.Text(), ""),
			IsError:    isError,
		}:
		}
	}
}

func (r *Runner) GetLogChannel() <-chan LogLine {
	return r.logChan
}

func (r *Runner) Stop() {
	r.cancel()

	for _, cmd := range r.processes {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	r.wg.Wait()
	close(r.logChan)
}

func (r *Runner) IsRunning() bool {
	for _, cmd := range r.processes {
		if cmd.Process != nil {
			return true
		}
	}
	return false
}

// Legacy function for backward compatibility
func Run(info ProjectInfo) error {
	if info.Type == TypeUnknown || len(info.Servers) == 0 {
		return fmt.Errorf("unable to detect project type")
	}

	runner := NewRunner()
	defer runner.Stop()

	if err := runner.Start(info); err != nil {
		return err
	}

	// Wait for context cancellation
	<-runner.ctx.Done()
	return nil
}
