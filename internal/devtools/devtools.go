package devtools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var DevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development tools",
	Long:  "Local development tools for project management",
}

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new project folder",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		if err := createProject(projectName); err != nil {
			fmt.Printf("Error creating project: %v\n", err)
			return
		}
		fmt.Printf("Project '%s' created successfully!\n", projectName)
	},
}

var serverCmd = &cobra.Command{
	Use:   "server [port]",
	Short: "Run a development server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port := args[0]
		fmt.Printf("Starting development server on port %s...\n", port)
		runDevServer(port)
	},
}

var boilerplateCmd = &cobra.Command{
	Use:   "boilerplate [type]",
	Short: "Generate boilerplate files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		boilerplateType := args[0]
		if err := generateBoilerplate(boilerplateType); err != nil {
			fmt.Printf("Error generating boilerplate: %v\n", err)
			return
		}
		fmt.Printf("Boilerplate '%s' generated successfully!\n", boilerplateType)
	},
}

func init() {
	DevCmd.AddCommand(createCmd)
	DevCmd.AddCommand(serverCmd)
	DevCmd.AddCommand(boilerplateCmd)
}

func createProject(name string) error {
	// Create main project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return err
	}

	// Create subdirectories
	dirs := []string{"src", "tests", "docs", "config"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(name, dir), 0755); err != nil {
			return err
		}
	}

	// Create README.md
	readmeContent := fmt.Sprintf("# %s\n\nProject description here.\n\n## Getting Started\n\n```bash\n# Install dependencies\nnpm install\n\n# Run development server\nnpm run dev\n```", name)

	if err := os.WriteFile(filepath.Join(name, "README.md"), []byte(readmeContent), 0644); err != nil {
		return err
	}

	// Create package.json for Node.js projects
	packageJSON := `{
  "name": "` + name + `",
  "version": "1.0.0",
  "description": "",
  "main": "index.js",
  "scripts": {
    "dev": "node src/index.js",
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "keywords": [],
  "author": "",
  "license": "ISC"
}`

	return os.WriteFile(filepath.Join(name, "package.json"), []byte(packageJSON), 0644)
}

func runDevServer(port string) {
	// Simple HTTP server implementation
	cmd := exec.Command("python3", "-m", "http.server", port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func generateBoilerplate(boilerplateType string) error {
	switch boilerplateType {
	case "web":
		return generateWebBoilerplate()
	case "api":
		return generateAPIBoilerplate()
	case "cli":
		return generateCLIBoilerplate()
	default:
		return fmt.Errorf("unsupported boilerplate type: %s", boilerplateType)
	}
}

func generateWebBoilerplate() error {
	files := map[string]string{
		"index.html": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My Web App</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <h1>Hello World!</h1>
    <script src="script.js"></script>
</body>
</html>`,
		"styles.css": `body {
    font-family: Arial, sans-serif;
    margin: 0;
    padding: 20px;
    background-color: #f5f5f5;
}

h1 {
    color: #333;
}`,
		"script.js": `console.log('Hello from JavaScript!');

document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM loaded');
});`,
	}

	for filename, content := range files {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func generateAPIBoilerplate() error {
	content := `from flask import Flask, jsonify, request

app = Flask(__name__)

@app.route('/')
def hello():
    return jsonify({'message': 'Hello World!'})

@app.route('/api/users', methods=['GET'])
def get_users():
    users = [
        {'id': 1, 'name': 'John Doe'},
        {'id': 2, 'name': 'Jane Smith'}
    ]
    return jsonify(users)

@app.route('/api/users', methods=['POST'])
def create_user():
    data = request.json
    return jsonify({'message': 'User created', 'data': data}), 201

if __name__ == '__main__':
    app.run(debug=True, port=5000)`

	return os.WriteFile("app.py", []byte(content), 0644)
}

func generateCLIBoilerplate() error {
	content := `#!/usr/bin/env python3
import argparse
import sys

def main():
    parser = argparse.ArgumentParser(description='My CLI Tool')
    parser.add_argument('--name', type=str, help='Your name')
    parser.add_argument('--verbose', '-v', action='store_true', help='Verbose output')
    
    args = parser.parse_args()
    
    if args.name:
        print(f"Hello, {args.name}!")
    else:
        print("Hello, World!")
    
    if args.verbose:
        print("Verbose mode enabled")

if __name__ == '__main__':
    main()`

	return os.WriteFile("cli.py", []byte(content), 0644)
}
