package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	cwd := "/app/workspace"

	filenames := []string{
		"test.txt",
		"foo/bar/test.txt",
		"../test.txt",
		"../../etc/passwd",
		"/etc/passwd",
	}

	for _, fname := range filenames {
		// Calculate the intended target path
        targetPath := filepath.Join(cwd, fname)

        // Ensure the resolved path remains inside the intended directory
        rel, err := filepath.Rel(cwd, targetPath)

        isSafe := err == nil && !strings.HasPrefix(rel, "..") && !strings.HasPrefix(rel, "/") && rel != ".."

        fmt.Printf("fname: %q\n", fname)
        fmt.Printf("targetPath: %q\n", targetPath)
        fmt.Printf("rel: %q (err: %v)\n", rel, err)
        fmt.Printf("isSafe: %v\n\n", isSafe)
	}
}
