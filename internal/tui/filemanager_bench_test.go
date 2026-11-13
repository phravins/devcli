package tui

import (
	"fmt"
	"testing"
)

func generateMockPaths(count int) []string {
	paths := make([]string, count)
	dirs := []string{"src", "cmd", "internal", "pkg", "assets", "vendor", "node_modules", "docs"}
	subdirs := []string{"tui", "utils", "server", "client", "models", "controllers", "views", "config"}
	exts := []string{".go", ".ts", ".js", ".json", ".md", ".py", ".yaml", ".html", ".css"}

	for i := 0; i < count; i++ {
		d := dirs[i%len(dirs)]
		sd := subdirs[(i/len(dirs))%len(subdirs)]
		ext := exts[i%len(exts)]
		paths[i] = fmt.Sprintf("/usr/local/project/%s/%s/file_%d_component%s", d, sd, i, ext)
	}
	return paths
}

func BenchmarkFilterFiles_Fuzzy(b *testing.B) {
	paths := generateMockPaths(10000)
	m := FileManagerModel{
		allFilePaths: paths,
		globalSearch: true,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m.filterFiles("component")
	}
}

func BenchmarkFilterFiles_FastPath(b *testing.B) {
	paths := generateMockPaths(25000)
	m := FileManagerModel{
		allFilePaths: paths,
		globalSearch: true,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m.filterFiles("component")
	}
}

func BenchmarkPerformSearchCmd(b *testing.B) {
	paths := generateMockPaths(20000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cmd := performSearchCmd(paths, "component")
		_ = cmd()
	}
}
