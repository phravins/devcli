package project

var Templates = map[string]map[string]string{
	"Go": {
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`,
		"go.mod": `module {{.Name}}

go 1.21
`,
	},
	"Python": {
		"main.py": `def main():
    print("Hello, World!")

if __name__ == "__main__":
    main()
`,
		"requirements.txt": `# Add your dependencies here
`,
	},
	"Node": {
		"package.json": `{
  "name": "{{.Name}}",
  "version": "1.0.0",
  "main": "index.js",
  "scripts": {
    "start": "node index.js"
  },
  "dependencies": {}
}
`,
		"index.js": `console.log("Hello, World!");
`,
	},
}
