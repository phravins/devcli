package snippets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Snippet struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Category    string    `json:"category"`
	Code        string    `json:"code"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Storage struct {
	filePath string
}

func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	devCLIDir := filepath.Join(homeDir, ".devcli")
	if err := os.MkdirAll(devCLIDir, 0755); err != nil {
		return nil, err
	}

	return &Storage{
		filePath: filepath.Join(devCLIDir, "snippets.json"),
	}, nil
}

func (s *Storage) LoadAll() ([]Snippet, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []Snippet{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var snippets []Snippet
	if err := json.Unmarshal(data, &snippets); err != nil {
		return nil, err
	}

	return snippets, nil
}

func (s *Storage) SaveAll(snippets []Snippet) error {
	data, err := json.MarshalIndent(snippets, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Storage) Add(snippet Snippet) error {
	snippets, err := s.LoadAll()
	if err != nil {
		return err
	}

	snippet.ID = generateID()
	snippet.CreatedAt = time.Now()
	snippet.UpdatedAt = time.Now()

	snippets = append(snippets, snippet)
	return s.SaveAll(snippets)
}

func (s *Storage) Update(snippet Snippet) error {
	snippets, err := s.LoadAll()
	if err != nil {
		return err
	}

	for i, snip := range snippets {
		if snip.ID == snippet.ID {
			snippet.UpdatedAt = time.Now()
			snippet.CreatedAt = snip.CreatedAt // Preserve original creation time
			snippets[i] = snippet
			return s.SaveAll(snippets)
		}
	}

	return fmt.Errorf("snippet not found")
}

func (s *Storage) Delete(id string) error {
	snippets, err := s.LoadAll()
	if err != nil {
		return err
	}

	for i, snip := range snippets {
		if snip.ID == id {
			snippets = append(snippets[:i], snippets[i+1:]...)
			return s.SaveAll(snippets)
		}
	}

	return fmt.Errorf("snippet not found")
}

func (s *Storage) Get(id string) (*Snippet, error) {
	snippets, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	for _, snip := range snippets {
		if snip.ID == id {
			return &snip, nil
		}
	}

	return nil, fmt.Errorf("snippet not found")
}

func (s *Storage) Search(query string) ([]Snippet, error) {
	all, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []Snippet

	for _, snip := range all {
		if strings.Contains(strings.ToLower(snip.Title), query) ||
			strings.Contains(strings.ToLower(snip.Description), query) ||
			strings.Contains(strings.ToLower(snip.Code), query) ||
			containsTag(snip.Tags, query) {
			results = append(results, snip)
		}
	}

	return results, nil
}

func (s *Storage) FilterByCategory(category string) ([]Snippet, error) {
	all, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	var results []Snippet
	for _, snip := range all {
		if strings.EqualFold(snip.Category, category) {
			results = append(results, snip)
		}
	}

	return results, nil
}

func (s *Storage) FilterByLanguage(language string) ([]Snippet, error) {
	all, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	var results []Snippet
	for _, snip := range all {
		if strings.EqualFold(snip.Language, language) {
			results = append(results, snip)
		}
	}

	return results, nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// GetDefaultSnippets returns some built-in snippets
func GetDefaultSnippets() []Snippet {
	return []Snippet{
		{
			ID:          "default-1",
			Title:       "HTTP Server (Go)",
			Description: "Basic HTTP server in Go",
			Language:    "go",
			Category:    "api",
			Code: `package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	
	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}`,
			Tags:      []string{"http", "server", "api"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "default-1-py",
			Title:       "HTTP Server (Python)",
			Description: "Basic HTTP server in Python (Flask)",
			Language:    "python",
			Category:    "api",
			Code: `from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return "Hello, World!"

if __name__ == '__main__':
    print("Server starting on :5000")
    app.run(port=5000, debug=True)`,
			Tags:      []string{"http", "server", "api", "flask"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "default-2",
			Title:       "Database Connection (Python)",
			Description: "PostgreSQL connection with psycopg2",
			Language:    "python",
			Category:    "database",
			Code: `import psycopg2

def get_db_connection():
    return psycopg2.connect(
        host="localhost",
        database="mydb",
        user="user",
        password="password"
    )

# Usage
conn = get_db_connection()
cursor = conn.cursor()
cursor.execute("SELECT * FROM users")
results = cursor.fetchall()
cursor.close()
conn.close()`,
			Tags:      []string{"database", "postgresql", "sql"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "default-3",
			Title:       "React Component",
			Description: "Functional React component with useState",
			Language:    "javascript",
			Category:    "ui",
			Code: `import React, { useState } from 'react';

function Counter() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <h1>Count: {count}</h1>
      <button onClick={() => setCount(count + 1)}>
        Increment
      </button>
      <button onClick={() => setCount(count - 1)}>
        Decrement
      </button>
    </div>
  );
}

export default Counter;`,
			Tags:      []string{"react", "component", "hooks"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}
