package fileops

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var FileCmd = &cobra.Command{
	Use:   "file",
	Short: "File operations",
	Long:  "File handling operations including read/write, copy/move, search, and processing",
}

var readCmd = &cobra.Command{
	Use:   "read [file]",
	Short: "Read a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		content, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}
		fmt.Print(string(content))
	},
}

var writeCmd = &cobra.Command{
	Use:   "write [file] [content]",
	Short: "Write to a file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		content := args[1]

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return
		}
		fmt.Printf("File '%s' written successfully!\n", file)
	},
}

var copyCmd = &cobra.Command{
	Use:   "copy [source] [destination]",
	Short: "Copy a file or directory",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		dest := args[1]

		if err := copyFile(source, dest); err != nil {
			fmt.Printf("Error copying: %v\n", err)
			return
		}
		fmt.Printf("Copied '%s' to '%s'\n", source, dest)
	},
}

var moveCmd = &cobra.Command{
	Use:   "move [source] [destination]",
	Short: "Move a file or directory",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		dest := args[1]

		if err := os.Rename(source, dest); err != nil {
			fmt.Printf("Error moving: %v\n", err)
			return
		}
		fmt.Printf("Moved '%s' to '%s'\n", source, dest)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [file]",
	Short: "Delete a file or directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]

		if err := os.RemoveAll(file); err != nil {
			fmt.Printf("Error deleting: %v\n", err)
			return
		}
		fmt.Printf("Deleted '%s'\n", file)
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [pattern] [directory]",
	Short: "Search for content in files",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := args[0]
		directory := args[1]

		matches, err := searchFiles(pattern, directory)
		if err != nil {
			fmt.Printf("Error searching: %v\n", err)
			return
		}

		for _, match := range matches {
			fmt.Println(match)
		}
	},
}

var jsonCmd = &cobra.Command{
	Use:   "json [action] [file]",
	Short: "Process JSON files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]

		switch action {
		case "format":
			if len(args) < 2 {
				fmt.Println("Please provide a JSON file")
				return
			}
			if err := formatJSON(args[1]); err != nil {
				fmt.Printf("Error formatting JSON: %v\n", err)
				return
			}
			fmt.Println("JSON formatted successfully!")
		case "validate":
			if len(args) < 2 {
				fmt.Println("Please provide a JSON file")
				return
			}
			if err := validateJSON(args[1]); err != nil {
				fmt.Printf("JSON validation failed: %v\n", err)
				return
			}
			fmt.Println("JSON is valid!")
		}
	},
}

var yamlCmd = &cobra.Command{
	Use:   "yaml [action] [file]",
	Short: "Process YAML files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]

		switch action {
		case "format":
			if len(args) < 2 {
				fmt.Println("Please provide a YAML file")
				return
			}
			if err := formatYAML(args[1]); err != nil {
				fmt.Printf("Error formatting YAML: %v\n", err)
				return
			}
			fmt.Println("YAML formatted successfully!")
		case "to-json":
			if len(args) < 2 {
				fmt.Println("Please provide a YAML file")
				return
			}
			if err := yamlToJSON(args[1]); err != nil {
				fmt.Printf("Error converting YAML to JSON: %v\n", err)
				return
			}
			fmt.Println("YAML converted to JSON!")
		}
	},
}

var renameCmd = &cobra.Command{
	Use:   "rename [pattern] [replacement] [directory]",
	Short: "Bulk rename files",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := args[0]
		replacement := args[1]
		directory := args[2]

		count, err := bulkRename(pattern, replacement, directory)
		if err != nil {
			fmt.Printf("Error renaming files: %v\n", err)
			return
		}
		fmt.Printf("Renamed %d files\n", count)
	},
}

func init() {
	FileCmd.AddCommand(readCmd)
	FileCmd.AddCommand(writeCmd)
	FileCmd.AddCommand(copyCmd)
	FileCmd.AddCommand(moveCmd)
	FileCmd.AddCommand(deleteCmd)
	FileCmd.AddCommand(searchCmd)
	FileCmd.AddCommand(jsonCmd)
	FileCmd.AddCommand(yamlCmd)
	FileCmd.AddCommand(renameCmd)
}

func copyFile(src, dst string) error {
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if sourceInfo.IsDir() {
		return copyDir(src, dst)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func searchFiles(pattern, directory string) ([]string, error) {
	var matches []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		if strings.Contains(string(content), pattern) {
			matches = append(matches, path)
		}

		return nil
	})

	return matches, err
}

func formatJSON(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, formatted, 0644)
}

func validateJSON(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var data interface{}
	return json.Unmarshal(content, &data)
}

func formatYAML(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var data interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return err
	}

	formatted, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, formatted, 0644)
}

func yamlToJSON(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var data interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return err
	}

	jsonContent, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	outputFile := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".json"
	return os.WriteFile(outputFile, jsonContent, 0644)
}

func bulkRename(pattern, replacement, directory string) (int, error) {
	count := 0

	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		oldName := entry.Name()
		if strings.Contains(oldName, pattern) {
			newName := strings.ReplaceAll(oldName, pattern, replacement)
			oldPath := filepath.Join(directory, oldName)
			newPath := filepath.Join(directory, newName)

			if err := os.Rename(oldPath, newPath); err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}
