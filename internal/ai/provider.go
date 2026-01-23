package ai

import "github.com/phravins/devcli/internal/config"

type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

type Provider interface {
	Name() string

	Model() string

	Configure(cfg *config.Config) error

	Send(messages []Message) (string, error)

	IsLocal() bool
}
