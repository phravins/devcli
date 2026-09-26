package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {

	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "testfile.txt")
	err := os.WriteFile(filePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !FileExists(filePath) {
		t.Errorf("FileExists(%q) returned false, expected true", filePath)
	}

	dirPath := filepath.Join(tempDir, "testdir")
	err = os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	if FileExists(dirPath) {
		t.Errorf("FileExists(%q) returned true for a directory, expected false", dirPath)
	}

	nonExistentPath := filepath.Join(tempDir, "doesnotexist.txt")
	if FileExists(nonExistentPath) {
		t.Errorf("FileExists(%q) returned true for non-existent path, expected false", nonExistentPath)
	}
}

func TestDirExists(t *testing.T) {

	tempDir := t.TempDir()

	dirPath := filepath.Join(tempDir, "testdir")
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	if !DirExists(dirPath) {
		t.Errorf("DirExists(%q) returned false, expected true", dirPath)
	}

	filePath := filepath.Join(tempDir, "testfile.txt")
	err = os.WriteFile(filePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if DirExists(filePath) {
		t.Errorf("DirExists(%q) returned true for a file, expected false", filePath)
	}

	nonExistentPath := filepath.Join(tempDir, "doesnotexist")
	if DirExists(nonExistentPath) {
		t.Errorf("DirExists(%q) returned true for non-existent path, expected false", nonExistentPath)
	}
}
