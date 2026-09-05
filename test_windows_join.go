package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	cwd := "C:\\app"

	filenames := []string{
		"C:\\Windows\\System32",
		"D:\\foo",
		"..\\..\\Windows",
	}

	for _, fname := range filenames {
		fmt.Printf("Join: %s\n", filepath.Join(cwd, fname))
	}
}
