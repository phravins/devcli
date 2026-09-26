package boilerplate

import (
	"fmt"
	"os"
	"path/filepath"
)

type Architecture struct {
	Name		string
	Description	string
	Structure	map[string]string
}

var Architectures = map[string]Architecture{

	"MVC": {
		Name:		"MVC",
		Description:	"Models, Views, Controllers layer separation",
		Structure: map[string]string{
			"models/":	"",
			"views/":	"",
			"controllers/":	"",
			"main.go": `package main

import "fmt"

func main() {
	// MVC Architecture Entry Point
	// Initialize your models, controllers, and views here.
	fmt.Println("Starting MVC App...")
}`,
			"models/README.md":		"Place your data models and database logic here.",
			"views/README.md":		"Place your templates and view logic here.",
			"controllers/README.md":	"Place your request handlers and business logic here.",
		},
	},

	"Modular": {
		Name:		"Modular Architecture",
		Description:	"Splits features into modules (auth, users, products, utils)",
		Structure: map[string]string{
			"auth/":	"",
			"users/":	"",
			"products/":	"",
			"utils/":	"",
			"main.go": `package main

import "fmt"

func main() {
	// Modular Architecture Entry Point
	// Register modules and start the server.
	fmt.Println("Starting Modular App...")
}`,
			"auth/README.md":	"Authentication module (Login, Signup, JWT).",
			"users/README.md":	"User management module.",
			"products/README.md":	"Product inventory module.",
		},
	},

	"Clean Architecture": {
		Name:		"Clean Architecture",
		Description:	"Domain, Application, Infrastructure, Presentation layers",
		Structure: map[string]string{
			"domain/":		"",
			"application/":		"",
			"infrastructure/":	"",
			"presentation/":	"",
			"main.go": `package main

import "fmt"

func main() {
	// Clean Architecture Entry Point
	// Wire up the layers: Infra -> Presentation -> App -> Domain
	fmt.Println("Starting Clean Architecture App...")
}`,
			"domain/README.md":		"Core business entities and interfaces.",
			"application/README.md":	"Use cases and application logic.",
			"infrastructure/README.md":	"External tools (DB, API clients).",
			"presentation/README.md":	"API handlers or CLI commands.",
		},
	},

	"Microservices": {
		Name:		"Microservices Layout",
		Description:	"Separated services workspace (Auth, Order, Notification)",
		Structure: map[string]string{
			"auth-service/":		"",
			"order-service/":		"",
			"notification-service/":	"",
			"gateway/":			"",
			"README.md":			"# Microservices Project\n\nThis workspace contains multiple independent services.\nEach folder is a self-contained service.",
			"auth-service/main.go": `package main
import "fmt"
// Auth Service Entry Point
func main() { fmt.Println("Starting Auth Service...") }`,
		},
	},
}

func (m *Manager) GenerateArchitecture(archName, destDir string) ([]string, error) {
	arch, ok := Architectures[archName]
	if !ok {
		return nil, fmt.Errorf("architecture '%s' not found", archName)
	}

	var createdFiles []string

	for path, content := range arch.Structure {
		fullPath := filepath.Join(destDir, path)
		if content == "" && (path[len(path)-1] == '/' || path[len(path)-1] == os.PathSeparator) {

			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return nil, err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				return nil, err
			}
			createdFiles = append(createdFiles, fullPath)
		}
	}
	return createdFiles, nil
}
