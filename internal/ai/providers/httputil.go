package providers

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// maxResponseBytes caps how much of a provider response is buffered.
	maxResponseBytes = 8 << 20
	// maxErrorBodyBytes caps how much of an error body is echoed to the user.
	maxErrorBodyBytes = 2 << 10
)

// cloudHTTPClient is shared by the hosted providers. The zero-value
// http.Client has no timeout, so a stalled connection would freeze the chat
// view indefinitely.
var cloudHTTPClient = &http.Client{Timeout: 120 * time.Second}

// redact removes every occurrence of secret from s.
func redact(s, secret string) string {
	if secret == "" {
		return s
	}
	return strings.ReplaceAll(s, secret, "[REDACTED]")
}

// redactErr returns err with secret scrubbed from its message, so transport
// errors (which can embed URLs or headers) never surface an API key.
func redactErr(err error, secret string) error {
	if err == nil || secret == "" || !strings.Contains(err.Error(), secret) {
		return err
	}
	return errors.New(redact(err.Error(), secret))
}

func truncateForError(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > maxErrorBodyBytes {
		return s[:maxErrorBodyBytes] + "…"
	}
	return s
}

// readErrorBody reads a bounded, redacted snippet of an error response body.
func readErrorBody(r io.Reader, secret string) string {
	b, _ := io.ReadAll(io.LimitReader(r, maxErrorBodyBytes+1))
	return truncateForError(redact(string(b), secret))
}
