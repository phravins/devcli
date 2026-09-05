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
		"C:\\Windows\\System32\\cmd.exe",
		"C:/Windows/System32/cmd.exe",
	}

	for _, fname := range filenames {
		// Clean and make it a relative path by force?
		// No, if fname is absolute, filepath.IsAbs(fname)

        target := filepath.Join(cwd, fname)

        // Ensure the target is still inside cwd
        // In Go, filepath.Join calls filepath.Clean, so target is clean.

        rel, err := filepath.Rel(cwd, target)

        fmt.Printf("fname: %q\n", fname)
        fmt.Printf("target: %q\n", target)
        fmt.Printf("rel: %q (err: %v)\n", rel, err)
        isSafe := !strings.HasPrefix(rel, "..") && !strings.HasPrefix(rel, "/") && rel != ".."
        fmt.Printf("isSafe: %v\n\n", isSafe)
	}
}
