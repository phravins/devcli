package devserver

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkDetect_FullStack(b *testing.B) {
	// Create a dummy project structure
	tmpDir := b.TempDir()

	backendDir := filepath.Join(tmpDir, "backend")
	frontendDir := filepath.Join(tmpDir, "frontend")
	os.MkdirAll(backendDir, 0755)
	os.MkdirAll(frontendDir, 0755)

	pkgJSON := `{
  "name": "dummy-project",
  "version": "1.0.0",
  "dependencies": {
    "lodash": "^4.17.21"
  }
}`

	os.WriteFile(filepath.Join(backendDir, "package.json"), []byte(pkgJSON), 0644)
	os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(pkgJSON), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Detect(tmpDir)
	}
}

func BenchmarkDetect_React(b *testing.B) {
	tmpDir := b.TempDir()

	pkgJSON := `{
  "name": "dummy-project",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.0.0"
  }
}`

	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Detect(tmpDir)
	}
}
