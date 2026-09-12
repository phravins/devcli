package tui

import (
	"fmt"
	"testing"
)

func generateBenchmarkPaths(count int) []string {
	paths := make([]string, count)
	dirs := []string{"src", "pkg", "cmd", "internal", "assets", "vendor", "docs", "tests", "build", "scripts"}
	subdirs := []string{"components", "utils", "models", "controllers", "views", "services", "config", "handlers", "middleware", "types"}
	exts := []string{".go", ".js", ".ts", ".py", ".json", ".md", ".yaml", ".html", ".css", ".sh"}

	for i := 0; i < count; i++ {
		d := dirs[i%len(dirs)]
		sd := subdirs[(i/len(dirs))%len(subdirs)]
		ext := exts[i%len(exts)]
		paths[i] = fmt.Sprintf("/home/user/projects/app_%d/%s/%s/file_%d%s", i%50, d, sd, i, ext)
	}
	return paths
}

func BenchmarkFilterFiles_Fuzzy_1000(b *testing.B) {
	paths := generateBenchmarkPaths(1000)
	m := FileManagerModel{
		allFilePaths: paths,
		globalSearch: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.filterFiles("file")
	}
}

func BenchmarkFilterFiles_Fuzzy_5000(b *testing.B) {
	paths := generateBenchmarkPaths(5000)
	m := FileManagerModel{
		allFilePaths: paths,
		globalSearch: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.filterFiles("file")
	}
}

func BenchmarkFilterFiles_Fuzzy_15000(b *testing.B) {
	paths := generateBenchmarkPaths(15000)
	m := FileManagerModel{
		allFilePaths: paths,
		globalSearch: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.filterFiles("file")
	}
}

func BenchmarkFilterFiles_FastPath_25000(b *testing.B) {
	paths := generateBenchmarkPaths(25000)
	m := FileManagerModel{
		allFilePaths: paths,
		globalSearch: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.filterFiles("file")
	}
}

func BenchmarkPerformSearchCmd_5000(b *testing.B) {
	paths := generateBenchmarkPaths(5000)
	cmd := performSearchCmd(paths, "file")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cmd()
	}
}

func BenchmarkPerformSearchCmd_15000(b *testing.B) {
	paths := generateBenchmarkPaths(15000)
	cmd := performSearchCmd(paths, "file")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cmd()
	}
}
