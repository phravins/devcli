package templates

// Template represents a project blueprint
type Template struct {
	Name        string
	Description string
	Stack       string // "Go", "Python", "Node", etc.
	Files       map[string]string
	InstallCmd  string //"npm install"
	RunCmd      string //"npm start"
}

// Registry holds the available templates
var Registry = []Template{
	{
		Name:        "Go Fiber API",
		Description: "High-performance Go web framework",
		Stack:       "Go",
		InstallCmd:  "go mod tidy",
		RunCmd:      "go run main.go",
		Files: map[string]string{
			"go.mod": `module {{.Name}}

go 1.21

require (
	github.com/gofiber/fiber/v2 v2.52.0
)
`,
			"main.go": `package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	log.Fatal(app.Listen(":3000"))
}
`,
			".gitignore": `bin/
.idea/
.vscode/
`,
		},
	},
	{
		Name:        "Python FastAPI",
		Description: "Modern, fast (high-performance) Python web framework",
		Stack:       "Python",
		InstallCmd:  "pip install -r requirements.txt",
		RunCmd:      "uvicorn main:app --reload",
		Files: map[string]string{
			"requirements.txt": `fastapi
uvicorn[standard]
`,
			"main.py": `from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"Hello": "World"}
`,
			".gitignore": `__pycache__/
*.pyc
.env
venv/
`,
		},
	},
	{
		Name:        "Node Express API",
		Description: "Fast, unopinionated, minimalist web framework for Node.js",
		Stack:       "Node",
		InstallCmd:  "npm install",
		RunCmd:      "npm start",
		Files: map[string]string{
			"package.json": `{
  "name": "{{.Name}}",
  "version": "1.0.0",
  "description": "",
  "main": "index.js",
  "scripts": {
    "start": "node index.js"
  },
  "dependencies": {
    "express": "^4.18.2"
  }
}
`,
			"index.js": `const express = require('express')
const app = express()
const port = 3000

app.get('/', (req, res) => {
  res.send('Hello World!')
})

app.listen(port, () => {
  console.log('Example app listening on port ' + port)
})
`,
			".gitignore": `node_modules/
.env
`,
		},
	},
	{
		Name:        "Java Console App",
		Description: "Standard Java application with Maven",
		Stack:       "Java",
		InstallCmd:  "mvn clean install",
		RunCmd:      "java -jar target/{{.Name}}-1.0-SNAPSHOT.jar",
		Files: map[string]string{
			"pom.xml": `<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example</groupId>
  <artifactId>{{.Name}}</artifactId>
  <packaging>jar</packaging>
  <version>1.0-SNAPSHOT</version>
  <name>{{.Name}}</name>
  <url>http://maven.apache.org</url>
  <properties>
    <maven.compiler.source>17</maven.compiler.source>
    <maven.compiler.target>17</maven.compiler.target>
  </properties>
  <dependencies>
    <dependency>
      <groupId>junit</groupId>
      <artifactId>junit</artifactId>
      <version>3.8.1</version>
      <scope>test</scope>
    </dependency>
  </dependencies>
  <build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-shade-plugin</artifactId>
            <version>3.2.4</version>
            <executions>
                <execution>
                    <phase>package</phase>
                    <goals>
                        <goal>shade</goal>
                    </goals>
                    <configuration>
                        <transformers>
                            <transformer implementation="org.apache.maven.plugins.shade.resource.ManifestResourceTransformer">
                                <mainClass>com.example.Main</mainClass>
                            </transformer>
                        </transformers>
                    </configuration>
                </execution>
            </executions>
        </plugin>
    </plugins>
  </build>
</project>
`,
			"src/main/java/com/example/Main.java": `package com.example;

public class Main {
    public static void main(String[] args) {
        System.out.println("Hello from {{.Name}}!");
    }
}
`,
			".gitignore": `target/
.idea/
*.iml
.vscode/
`,
		},
	},
	{
		Name:        "Kotlin Console App",
		Description: "Kotlin application with Gradle (Kotlin DSL)",
		Stack:       "Kotlin",
		InstallCmd:  "gradle build",
		RunCmd:      "gradle run",
		Files: map[string]string{
			"build.gradle.kts": `plugins {
    kotlin("jvm") version "1.9.20"
    application
}

group = "com.example"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {
    testImplementation(kotlin("test"))
}

application {
    mainClass.set("MainKt")
}

tasks.test {
    useJUnitPlatform()
}
`,
			"settings.gradle.kts": `rootProject.name = "{{.Name}}"
`,
			"src/main/kotlin/Main.kt": `fun main() {
    println("Hello from {{.Name}}!")
}
`,
			".gitignore": `.gradle/
build/
.idea/
*.iml
`,
			// We can't easily generate the gradlew wrapper binary scripts here without embedding them.
			// For now, we assume user has gradle installed or we can simplify to not rely on wrapper
			// IF we want to be safe, we change InstallCmd to "gradle build"
			"README.md": `# {{.Name}}

To build and run:
` + "```bash" + `
gradle build
gradle run
` + "```" + `
`,
		},
	},
	{
		Name:        "Dart Console App",
		Description: "Command-line application using Dart SDK",
		Stack:       "Dart",
		InstallCmd:  "dart pub get",
		RunCmd:      "dart run",
		Files: map[string]string{
			"pubspec.yaml": `name: {{.Name}}
description: A sample command-line application.
version: 1.0.0
environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  path: ^1.8.0

dev_dependencies:
  lints: ^2.0.0
  test: ^1.21.0
`,
			"bin/main.dart": `import 'package:{{.Name}}/{{.Name}}.dart' as app;

void main(List<String> arguments) {
  print('Hello world: ${app.calculate()}!');
}
`,
			// We need a proper library file structure for standard dart recommended layout
			"lib/{{.Name}}.dart": `int calculate() {
  return 6 * 7;
}
`,
			".gitignore": `.dart_tool/
build/
pubspec.lock
.packages
.vscode/
`,
			"analysis_options.yaml": `include: package:lints/recommended.yaml
`,
		},
	},
	{
		Name:        "C++ Console App",
		Description: "C++ Standard Application with CMake",
		Stack:       "C++",
		InstallCmd:  "cmake -B build -S . && cmake --build build",
		RunCmd:      "./build/{{.Name}}", // On Windows this might need .exe handling but keeping simple for now
		Files: map[string]string{
			"CMakeLists.txt": `cmake_minimum_required(VERSION 3.10)

project({{.Name}} VERSION 1.0)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED True)

add_executable({{.Name}} src/main.cpp)
`,
			"src/main.cpp": `#include <iostream>

int main() {
    std::cout << "Hello from {{.Name}}!" << std::endl;
    return 0;
}
`,
			".gitignore": `build/
.vscode/
.idea/
`,
		},
	},
}

// Get returns the full list (helper for external access)
func List() []Template {
	return Registry
}
