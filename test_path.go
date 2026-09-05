package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	cwd := "/app"

	filenames := []string{
		"/etc/passwd",
		"C:\\Windows",
		"../foo",
	}

	for _, fname := range filenames {
		fmt.Printf("Join: %s\n", filepath.Join(cwd, fname))
	}
}
