package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func StripExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

func PrintSuccess(msg string) {
	fmt.Printf("\033[32m %s\033[0m\n", msg)
}

func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "\033[31m %s\033[0m\n", msg)
}

func FindExecutable(cmdName string, fallbackGlobs []string) string {
	if path, err := exec.LookPath(cmdName); err == nil {
		return path
	}

	for _, pattern := range fallbackGlobs {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {

			return matches[0]
		}
	}

	return ""
}

func DeepSearchExecutable(cmdName string, roots []string) string {

	for _, root := range roots {

		found := ""
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return filepath.SkipDir
			}

			rel, err := filepath.Rel(root, path)
			if err != nil {
				return filepath.SkipDir
			}
			depth := strings.Count(rel, string(os.PathSeparator))
			if depth > 3 {
				return filepath.SkipDir
			}

			if !info.IsDir() && (info.Name() == cmdName || info.Name() == cmdName+".exe") {
				found = path
				return filepath.SkipAll
			}
			return nil
		})

		if found != "" {
			return found
		}
	}
	return ""
}
