package providers

import (
	"fmt"
	"strings"
	"testing"
	"github.com/phravins/devcli/internal/ai"
)

func BenchmarkLocalHFProvider_StringConcat(b *testing.B) {
	messages := make([]ai.Message, 1000)
	for i := 0; i < 1000; i++ {
		messages[i] = ai.Message{
			Role:    "user",
			Content: "This is a slightly longer test message for testing string concatenation efficiency.",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var prompt string
		for _, m := range messages {
			prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
		}
		prompt += "assistant:"
		_ = prompt
	}
}

func BenchmarkLocalHFProvider_StringsBuilder(b *testing.B) {
	messages := make([]ai.Message, 1000)
	for i := 0; i < 1000; i++ {
		messages[i] = ai.Message{
			Role:    "user",
			Content: "This is a slightly longer test message for testing string concatenation efficiency.",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		for _, m := range messages {
			builder.WriteString(m.Role)
			builder.WriteString(": ")
			builder.WriteString(m.Content)
			builder.WriteString("\n")
		}
		builder.WriteString("assistant:")
		prompt := builder.String()
		_ = prompt
	}
}
