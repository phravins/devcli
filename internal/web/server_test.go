package web

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{"simple", "ls -la", []string{"ls", "-la"}},
		{"single quotes", "echo 'hello world'", []string{"echo", "hello world"}},
		{"double quotes", "echo \"hello world\"", []string{"echo", "hello world"}},
		{"escaped spaces", `cat file\ with\ space.txt`, []string{"cat", "file with space.txt"}},
		{"multiple spaces", "npm   run   dev", []string{"npm", "run", "dev"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommand(tt.command)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseCommand(%q) = %v, want %v", tt.command, got, tt.expected)
			}
		})
	}
}

func TestRunShellAllowed(t *testing.T) {
	_, err := runShell("ls")
	if err != nil {
		t.Errorf("expected ls to be allowed, got error: %v", err)
	}
}

func TestRunShellDisallowed(t *testing.T) {
	_, err := runShell("rm -rf /")
	if err == nil {
		t.Errorf("expected rm to be disallowed, but it succeeded")
	}
}
