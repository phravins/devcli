package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileExists returns true if the given path exists and is a file
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirExists returns true if the given path exists and is a directory
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// EnsureDir creates a directory (and any parents) if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// StripExt returns the file name without extension
func StripExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// JoinPath is a convenience wrapper around filepath.Join
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// PrintSuccess prints a green success message
func PrintSuccess(msg string) {
	fmt.Printf("\033[32m %s\033[0m\n", msg)
}

// PrintError prints a red error message
func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "\033[31m %s\033[0m\n", msg)
}

// FindExecutable attempts to find an executable by name in PATH or fallback glob patterns.
func FindExecutable(cmdName string, fallbackGlobs []string) string {
	// 1. Try PATH
	if path, err := exec.LookPath(cmdName); err == nil {
		return path
	}

	// 2. Try Fallbacks
	for _, pattern := range fallbackGlobs {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			// Return the first match found
			return matches[0]
		}
	}

	return ""
}

// DeepSearchExecutable performs a more intensive search in specific root directories.
// It looks for the cmdName in subdirectories of roots, but limits depth for performance.
func DeepSearchExecutable(cmdName string, roots []string) string {
	// Common patterns for compilers to narrow down the search
	// e.g. for "gcc" we might look for folders containing "mingw", "codeblocks", etc.
	for _, root := range roots {
		// We walk only 3 levels deep to avoid scanning the entire disk
		found := ""
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return filepath.SkipDir
			}

			// Optimization: Skip very deep directories
			rel, _ := filepath.Rel(root, path)
			depth := strings.Count(rel, string(os.PathSeparator))
			if depth > 3 {
				return filepath.SkipDir
			}

			if !info.IsDir() && (info.Name() == cmdName || info.Name() == cmdName+".exe") {
				found = path
				return filepath.SkipAll // Stop walking
			}
			return nil
		})

		if found != "" {
			return found
		}
	}
	return ""
}
