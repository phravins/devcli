package tui

import (
	"github.com/phravins/devcli/internal/updater"
	"testing"
)

func BenchmarkUpdaterModelView(b *testing.B) {
	m := NewUpdaterModel()
	m.width = 80
	m.height = 24
	m.info = &updater.UpdateInfo{
		CurrentVersion:    "1.0.0",
		LatestVersion:     "1.1.0",
		IsUpdateAvailable: true,
		ReleaseNotes:      "Release notes for 1.1.0.\n- Feature 1\n- Fix 2\n- Optimization 3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.View()
	}
}
