package devtools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Container struct {
	ID        string `json:"ID"`
	Image     string `json:"Image"`
	Command   string `json:"Command"`
	CreatedAt string `json:"CreatedAt"`
	Status    string `json:"Status"`
	Names     string `json:"Names"`
	State     string `json:"State"`
}

// IsDockerAvailable checks if docker command is installed and daemon is running
func IsDockerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, "docker", "info").Run()
	return err == nil
}

// ListContainers returns formatted list of docker containers
func ListContainers() ([]Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", "{{json .}}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker ps failed: %s", string(out))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var containers []Container

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var c Container
		if err := json.Unmarshal([]byte(line), &c); err == nil {
			containers = append(containers, c)
		}
	}

	return containers, nil
}

// StartContainer starts a docker container
func StartContainer(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "start", id).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(out))
	}
	return nil
}

// StopContainer stops a docker container
func StopContainer(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "stop", id).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(out))
	}
	return nil
}

// RestartContainer restarts a docker container
func RestartContainer(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "restart", id).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(out))
	}
	return nil
}

// GetContainerLogs retrieves logs from container
func GetContainerLogs(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "logs", "--tail", "100", id).CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}
